/*
 * Copyright 2020-present Open Networking Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	p4confv1 "github.com/p4lang/p4runtime/go/p4/config/v1"
	p4v1 "github.com/p4lang/p4runtime/go/p4/v1"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"io/ioutil"
	"mapr/fabric"
	"mapr/translate"
	"net"
	"strings"
)

var (
	target p4v1.P4RuntimeClient
	port   = flag.Int("port", 28001,
		"The server port")
	targetAddr = flag.String("target_addr", "127.0.0.1:28000",
		"The target address in the format of host:port")
	processorName = flag.String("proc", "dummy",
		"Processor to use")
	logicalP4InfoPath = flag.String("logical_p4info", "",
		"Path to logical P4Info file in binary format, e.g., `p4info.bin`")
	targetP4ConfigPaths = flag.String("target_p4_config", "",
		"Path to P4 pipeline config files to apply to target, e.g., `p4info.bin,bmv2.json`")
)

const MaxMsgLen = 255

type MsgDirection string

const (
	FromCtrl   MsgDirection = "ctrl >>"
	ToCtrl     MsgDirection = "ctrl <<"
	FromTarget MsgDirection = "<< trgt"
	ToTarget   MsgDirection = ">> trgt"
)

func logMsg(dir MsgDirection, msg proto.Message) {
	msgString := proto.CompactTextString(msg)
	msgLen := len(msgString)
	if msgLen > MaxMsgLen {
		msgString = msgString[:MaxMsgLen] + fmt.Sprintf("... truncated %d bytes", msgLen-MaxMsgLen)
	}
	log.WithField("proto", msgString).Debugf("%s %T", dir, msg)
}

type Server struct {
	// Holds the logical P4RT entities.
	P4RtStore translate.P4RtStore
	// Handles translation of logical updates to physical ones.
	Translator translate.Translator
}

func NewServer() *Server {
	ctx := translate.NewContext()
	var trn translate.Translator
	if *processorName == "dummy" {
		trn = translate.NewDummyTranslator()
	} else {
		var proc translate.Processor
		switch *processorName {
		case "fabric":
			proc = fabric.NewFabricProcessor(ctx)
		default:
			panic("Unknown processor")
		}
		trn = translate.NewTranslator(proc, ctx)
	}
	return &Server{
		P4RtStore:  translate.NewP4RtStore("logical"),
		Translator: trn,
	}
}

func (s Server) Capabilities(ctx context.Context, request *p4v1.CapabilitiesRequest) (*p4v1.CapabilitiesResponse, error) {
	logMsg(FromCtrl, request)
	response, err := target.Capabilities(ctx, request)
	if err != nil {
		return nil, err
	}
	logMsg(ToCtrl, response)
	return response, nil
}

var globalWriteCount = 1

func (s Server) Write(ctx context.Context, logicalReq *p4v1.WriteRequest) (*p4v1.WriteResponse, error) {
	writeCount := globalWriteCount
	globalWriteCount++
	log.Debugf("@@@@@@ BEGIN WRITE REQUEST #%d @@@@@@ ", writeCount)
	defer log.Debugf("@@@@@@ END WRITE REQUEST #%d @@@@@@", writeCount)

	logMsg(FromCtrl, logicalReq)

	if logicalReq.Atomicity != p4v1.WriteRequest_CONTINUE_ON_ERROR {
		return nil, status.Errorf(codes.Unimplemented, "Atomicity should be CONTINUE_ON_ERROR")
	}

	// Template WriteRequest for the target.
	// We'll emit one or none for each Update in the logical request.
	physicalRequest := p4v1.WriteRequest{
		DeviceId:   logicalReq.DeviceId,
		RoleId:     logicalReq.RoleId,
		ElectionId: logicalReq.ElectionId,
		Updates:    nil,
		Atomicity:  logicalReq.Atomicity,
	}

	ok := true
	for _, logicalUpdate := range logicalReq.Updates {
		// Validate update against P4RT store, to catch duplicate entries, and other P4RT-level errors.
		if err := s.P4RtStore.ApplyUpdate(logicalUpdate, true); err != nil {
			log.Errorf("ServerStore.ApplyUpdate(dry_run=true): %v [%v]", err, logicalUpdate)
			ok = false
			continue // next update
		}

		// Translate logical update to zero or more physical ones to write on the target.
		targetUpdates, err := s.Translator.Translate(logicalUpdate)
		if err != nil {
			log.Errorf("Translator.Translate(): %v [%v]", err, logicalUpdate)
			ok = false
			continue // next update
		}

		if targetUpdates != nil && len(targetUpdates) > 0 {
			// Write physical updates to target.
			physicalRequest.Updates = targetUpdates
			logMsg(ToTarget, &physicalRequest)
			_, err = target.Write(ctx, &physicalRequest)
			if err != nil {
				// TODO (carmelo): unpack and log P4RT error trailers from target.
				log.Errorf("%s %v", FromTarget, err)
				ok = false
				continue // next update
			}
			// Write RPC was successful!
		}

		// Update internal stores (there should be no errors since we did a dry run before)
		if err := s.P4RtStore.ApplyUpdate(logicalUpdate, false); err != nil {
			panic(err)
		}
		if err := s.Translator.ApplyUpdate(logicalUpdate, targetUpdates); err != nil {
			panic(err)
		}
	}

	// Send WriteResponse or error to controller.
	if ok {
		response := &p4v1.WriteResponse{}
		logMsg(ToCtrl, response)
		return response, nil
	} else {
		// FIXME (carmelo): return errors compliant with P4RT spec. I.e., append trailers with details for each update
		//  on the logical WriteRequest.
		return nil, status.Errorf(codes.Unknown, "Check mapr.log")
	}
}

func (s Server) Read(request *p4v1.ReadRequest, toClient p4v1.P4Runtime_ReadServer) error {
	// TODO: read from store, not from target
	logMsg(FromCtrl, request)
	ctx, cancel := context.WithCancel(toClient.Context())
	defer cancel()
	fromTarget, err := target.Read(ctx, request)
	if err != nil {
		return err
	}
	for {
		response, err := fromTarget.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		logMsg(ToCtrl, response)
		if err := toClient.Send(response); err != nil {
			return err
		}
	}
}

func (s Server) SetForwardingPipelineConfig(ctx context.Context, request *p4v1.SetForwardingPipelineConfigRequest) (
	*p4v1.SetForwardingPipelineConfigResponse, error) {
	logMsg(FromCtrl, request)
	// Compare P4Info in request with the one passed via flags. Return error if not equal.
	if bytes, err := ioutil.ReadFile(*logicalP4InfoPath); err == nil {
		logicalP4Info := &p4confv1.P4Info{}
		if err := proto.Unmarshal(bytes, logicalP4Info); err != nil {
			panic(err)
		}
		if !proto.Equal(request.Config.P4Info, logicalP4Info) {
			return nil, status.Error(codes.InvalidArgument, "mapr: P4Info not supported")
		}
	} else {
		panic(err)
	}
	// Modify request by swapping config with target one
	pieces := strings.Split(*targetP4ConfigPaths, ",")
	// Read and parse physical p4info.bin
	if bytes, err := ioutil.ReadFile(pieces[0]); err == nil {
		if err := proto.Unmarshal(bytes, request.Config.P4Info); err != nil {
			panic(err)
		}
	} else {
		panic(err)
	}
	// Read P4 config blob
	if bytes, err := ioutil.ReadFile(pieces[1]); err == nil {
		request.Config.P4DeviceConfig = bytes
	} else {
		panic(err)
	}
	// Forward modified request
	logMsg(ToTarget, request)
	if response, err := target.SetForwardingPipelineConfig(ctx, request); err == nil {
		logMsg(ToCtrl, response)
		return response, nil
	} else {
		return nil, err
	}
}

func (s Server) GetForwardingPipelineConfig(ctx context.Context, request *p4v1.GetForwardingPipelineConfigRequest) (
	// TODO: implement returning logical config instead of physical one
	*p4v1.GetForwardingPipelineConfigResponse, error) {
	logMsg(FromCtrl, request)
	response, err := target.GetForwardingPipelineConfig(ctx, request)
	if err != nil {
		return nil, err
	}
	logMsg(ToCtrl, response)
	return response, nil
}

func (s Server) StreamChannel(inStream p4v1.P4Runtime_StreamChannelServer) error {
	log.Println("StreamChannel opened!")
	defer log.Println("StreamChannel closed!")

	outCtx, outCancel := context.WithCancel(inStream.Context())
	defer outCancel()
	outStream, err := target.StreamChannel(outCtx)
	if err != nil {
		return err
	}

	waiterr := make(chan error)

	go func() {
		for {
			response, err := outStream.Recv()
			if err != nil {
				waiterr <- err
				return
			}
			logMsg(FromTarget, response)
			if err := inStream.Send(response); err != nil {
				waiterr <- err
				return
			}
		}
	}()

	go func() {
		for {
			request, err := inStream.Recv()
			if err != nil {
				if err == io.EOF {
					err = outStream.CloseSend()
				}
				waiterr <- err
				return
			}
			logMsg(FromCtrl, request)
			if err := outStream.Send(request); err != nil {
				waiterr <- err
				return
			}
		}
	}()

	if err := <-waiterr; err == nil || err == io.EOF {
		return nil
	} else {
		return err
	}
}

func Start(port int, targetAddr string) {
	// Client to target
	conn, err := grpc.Dial(targetAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial target: %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()
	target = p4v1.NewP4RuntimeClient(conn)

	// Server
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	server := grpc.NewServer()
	p4v1.RegisterP4RuntimeServer(server, NewServer())
	log.Printf("Listening for controller on port %d, talking to target on %s...\n", port, targetAddr)
	_ = server.Serve(lis)
}

func main() {
	flag.Parse()
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
		DisableQuote:  true})
	log.SetLevel(log.TraceLevel)
	Start(*port, *targetAddr)
}
