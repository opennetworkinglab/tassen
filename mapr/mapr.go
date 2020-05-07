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
	"github.com/p4lang/p4runtime/go/p4/v1"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
)

var (
	client     v1.P4RuntimeClient
	port       = flag.Int("port", 28001, "The server port")
	targetAddr = flag.String("target_addr", "127.0.0.1:28000", "The target address in the format of host:port")
)

const MaxMsgLen = 255

func logMsg(dir string, msg fmt.Stringer) {
	msgString := msg.String()
	msgLen := len(msgString)
	truncString := ""
	if msgLen > MaxMsgLen {
		msgString = msgString[:MaxMsgLen]
		truncString = fmt.Sprintf("... truncated %d bytes", msgLen-MaxMsgLen)
	}
	log.Printf("%s %T { %s%s }\n", dir, msg, msgString, truncString)
}

type server struct {
}

func (p server) Capabilities(ctx context.Context, request *v1.CapabilitiesRequest) (*v1.CapabilitiesResponse, error) {
	logMsg("<<", request)
	response, err := client.Capabilities(ctx, request)
	if err != nil {
		return nil, err
	}
	logMsg(">>", response)
	return response, nil
}

func (p server) Write(ctx context.Context, request *v1.WriteRequest) (*v1.WriteResponse, error) {
	logMsg("<<", request)
	response, err := client.Write(ctx, request)
	if err != nil {
		return nil, err
	}
	logMsg(">>", response)
	return response, nil
}

func (p server) Read(request *v1.ReadRequest, toClient v1.P4Runtime_ReadServer) error {
	logMsg("<<", request)
	ctx, cancel := context.WithCancel(toClient.Context())
	defer cancel()
	fromTarget, err := client.Read(ctx, request)
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
		logMsg(">>", response)
		if err := toClient.Send(response); err != nil {
			return err
		}
	}
}

func (p server) SetForwardingPipelineConfig(ctx context.Context, request *v1.SetForwardingPipelineConfigRequest) (
	*v1.SetForwardingPipelineConfigResponse, error) {
	logMsg("<<", request)
	response, err := client.SetForwardingPipelineConfig(ctx, request)
	if err != nil {
		return nil, err
	}
	logMsg(">>", response)
	return response, nil
}

func (p server) GetForwardingPipelineConfig(ctx context.Context, request *v1.GetForwardingPipelineConfigRequest) (
	*v1.GetForwardingPipelineConfigResponse, error) {
	logMsg("<<", request)
	response, err := client.GetForwardingPipelineConfig(ctx, request)
	if err != nil {
		return nil, err
	}
	logMsg(">>", response)
	return response, nil
}

func (p server) StreamChannel(inStream v1.P4Runtime_StreamChannelServer) error {
	log.Println("StreamChannel opened!")
	defer log.Println("StreamChannel closed!")

	outCtx, outCancel := context.WithCancel(inStream.Context())
	defer outCancel()
	outStream, err := client.StreamChannel(outCtx)
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
			logMsg(">>", response)
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
			logMsg("<<", request)
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

func newServer() *server {
	s := &server{}
	return s
}

func Start(port int, targetAddr string) {
	// Client
	conn, err := grpc.Dial(targetAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial target: %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()
	client = v1.NewP4RuntimeClient(conn)

	// Server
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	server := grpc.NewServer()
	v1.RegisterP4RuntimeServer(server, newServer())
	log.Printf("Listening on port %d, talking to %s...\n", port, targetAddr)
	_ = server.Serve(lis)
}

func main() {
	flag.Parse()

	Start(*port, *targetAddr)
}
