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
	"mapr/store"
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
	translatorName = flag.String("translator", "dummy",
		"Translator to use")
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

type server struct {
	translator  translate.Translator
	serverStore store.P4RtStore
	targetStore store.P4RtStore
	tassenStore store.TassenStore
}

func newServer() *server {
	serverStore := store.NewP4RtStore()
	targetStore := store.NewP4RtStore()
	tassenStore := store.NewTassenStore()
	var translator translate.Translator
	switch *translatorName {
	case "dummy":
		translator = translate.NewDummyTranslator()
	case "fabric":
		translator = translate.NewFabricTranslator(serverStore, targetStore, tassenStore)
	default:
		panic("Unknown translator")
	}
	return &server{
		translator:  translator,
		serverStore: serverStore,
		targetStore: targetStore,
		tassenStore: tassenStore,
	}
}

func (s server) Capabilities(ctx context.Context, request *p4v1.CapabilitiesRequest) (*p4v1.CapabilitiesResponse, error) {
	logMsg(FromCtrl, request)
	response, err := target.Capabilities(ctx, request)
	if err != nil {
		return nil, err
	}
	logMsg(ToCtrl, response)
	return response, nil
}

func (s server) Write(ctx context.Context, logicalReq *p4v1.WriteRequest) (*p4v1.WriteResponse, error) {
	logMsg(FromCtrl, logicalReq)
	// Validate logical request by executing an update dry run in the stores.
	if err := s.serverStore.Update(logicalReq, true); err != nil {
		return nil, err
	}
	if err := s.tassenStore.Update(logicalReq, true); err != nil {
		return nil, err
	}
	// Translate request to one that can be applied to the physical pipeline on the target.
	physicalReq, err := s.translator.Translate(logicalReq)
	if err != nil {
		return nil, err
	}
	var response *p4v1.WriteResponse = nil
	if physicalReq != nil {
		logMsg(ToTarget, physicalReq)
		// Translator wants to update the target.
		// Validate request.
		if err := s.targetStore.Update(physicalReq, true); err != nil {
			return nil, err
		}
		// Apply on target.
		res, err := target.Write(ctx, physicalReq)
		if err != nil {
			return nil, err
		}
		response = res
		// Store.
		_ = s.targetStore.Update(physicalReq, false)
	} else {
		// No need to update the target.
		// Fake successful response.
		response = &p4v1.WriteResponse{}
	}
	logMsg(ToCtrl, response)
	// Target updated successfully, update store with logical entities.
	_ = s.serverStore.Update(logicalReq, false)
	_ = s.tassenStore.Update(logicalReq, false)
	return response, nil
}

func (s server) Read(request *p4v1.ReadRequest, toClient p4v1.P4Runtime_ReadServer) error {
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

func (s server) SetForwardingPipelineConfig(ctx context.Context, request *p4v1.SetForwardingPipelineConfigRequest) (
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

func (s server) GetForwardingPipelineConfig(ctx context.Context, request *p4v1.GetForwardingPipelineConfigRequest) (
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

func (s server) StreamChannel(inStream p4v1.P4Runtime_StreamChannelServer) error {
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
	p4v1.RegisterP4RuntimeServer(server, newServer())
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
