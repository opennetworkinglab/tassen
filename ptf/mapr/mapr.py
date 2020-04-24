#!/usr/bin/env python2

# Copyright 2020-present Open Networking Foundation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

import argparse
import time
import traceback
from functools import wraps

import grpc
from concurrent import futures
from google.protobuf import text_format
from p4.v1 import p4runtime_pb2_grpc

DEFAULT_TIMEOUT = 10


def explained_msg(msg):
    """
    Returns a string explaining the given Protobuf message.
    """
    name = msg.__class__.__name__
    oneline = text_format.MessageToString(msg, as_one_line=True)
    if len(oneline) > 255:
        oneline = oneline[0:255] + '... TRUNCATED!'
    return "%s { %s }" % (name, oneline)


def explained_error(grpc_error):
    """
    Returns a string explaining the given gRPC error.
    """
    assert isinstance(grpc_error, grpc.RpcError)
    return "%s: %s" % (grpc_error.code(), grpc_error.details())


def logged_msg_iterator(iterator, fmt):
    """
    Returns an iterator that logs all messages in the given one using the given format.
    """
    for msg in iterator:
        print(fmt % explained_msg(msg))
        yield msg


def logged_unary(f):
    """
    Servicer method annotation logs request and response for unary RPCs.
    """

    @wraps(f)
    def handle(*args, **kwargs):
        assert isinstance(args[0], RelayP4RuntimeServicer)
        request = args[1]
        print(">> %s" % explained_msg(request))
        response = f(*args, **kwargs)
        print("<< %s " % explained_msg(response))
        return response

    return handle


def logged_server_stream(f):
    """
    Servicer method annotation logs request and responses for server stream RPCs.
    """

    @wraps(f)
    def handle(*args, **kwargs):
        assert isinstance(args[0], RelayP4RuntimeServicer)
        request = args[1]
        print(">> %s" % explained_msg(request))
        response = f(*args, **kwargs)
        return logged_msg_iterator(response, "<< %s ")

    return handle


def logged_bidi_stream(f):
    """
    Servicer method annotation logs requests and responses for bidirectional stream RPCs.
    """

    @wraps(f)
    def handle(*args, **kwargs):
        assert isinstance(args[0], RelayP4RuntimeServicer)
        # Replace request iterator in args with a logged one
        new_args = list(args)
        new_args[1] = logged_msg_iterator(args[1], '<< %s')
        # Return logged response iterator
        return logged_msg_iterator(f(*new_args, **kwargs), '>> %s')

    return handle


def relay_rpc_errors(f):
    """
    Servicer method annotation relays gRPC errors from the target.
    """

    @wraps(f)
    def handle(*args, **kwargs):
        assert isinstance(args[0], RelayP4RuntimeServicer)
        assert isinstance(args[2], grpc.ServicerContext)
        context = args[2]
        try:
            return f(*args, **kwargs)
        except grpc.RpcError as e:
            print("<< %s" % explained_error(e))
            context.set_code(e.code())
            context.set_details(e.details())
        except Exception as ex:
            traceback.print_exc()
            context.set_code(grpc.StatusCode.UNKNOWN)
            context.set_details("Mapper error: %s" % ex.message)

    return handle


class RelayP4RuntimeServicer(p4runtime_pb2_grpc.P4RuntimeServicer):
    """
    Implementation of a P4Runtime servicer that relays all calls to a given target.
    """

    def __init__(self, channel):
        self.stub = p4runtime_pb2_grpc.P4RuntimeStub(channel)

    @relay_rpc_errors
    @logged_unary
    def Write(self, request, context):
        return self.stub.Write(request, timeout=DEFAULT_TIMEOUT)

    @relay_rpc_errors
    @logged_server_stream
    def Read(self, request, context):
        return self.stub.Read(request, timeout=DEFAULT_TIMEOUT)

    @relay_rpc_errors
    @logged_unary
    def SetForwardingPipelineConfig(self, request, context):
        return self.stub.SetForwardingPipelineConfig(request, timeout=DEFAULT_TIMEOUT)

    @relay_rpc_errors
    @logged_unary
    def GetForwardingPipelineConfig(self, request, context):
        return self.stub.GetForwardingPipelineConfig(request, timeout=DEFAULT_TIMEOUT)

    @relay_rpc_errors
    @logged_bidi_stream
    def StreamChannel(self, request_iterator, context):
        return self.stub.StreamChannel(request_iterator)


def mapr_start(target_addr, server_port):
    """
    Starts the mapr server.
    """
    # create a gRPC server
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=5))
    channel = grpc.insecure_channel(target_addr)
    p4runtime_pb2_grpc.add_P4RuntimeServicer_to_server(RelayP4RuntimeServicer(channel), server)

    # listen on port 50051
    print('Starting mapr server. Listening on port %d...' % server_port)
    server.add_insecure_port('[::]:%d' % server_port)
    server.start()
    return server


def parse_args():
    parser = argparse.ArgumentParser(description="P4Runtime mapper")
    parser.add_argument(
        '--server-port',
        help='Port where the mapr server will listen for P4Runtime RPCs',
        type=int,
        default='28001',
    )
    parser.add_argument(
        '--target-addr',
        help='Address of the target P4Runtime server',
        type=str,
        default='localhost:28000',
    )
    return parser.parse_args()


def main():
    args = parse_args()
    server = mapr_start(args.target_addr, args.server_port)
    # since server.start() will not block,
    # a sleep-loop is added to keep alive
    try:
        while True:
            time.sleep(86400)
    except KeyboardInterrupt:
        server.stop(0)


if __name__ == '__main__':
    main()
