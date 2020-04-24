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

import unittest
from functools import wraps

import grpc
from concurrent import futures
from p4.v1 import p4runtime_pb2_grpc
from p4.v1.p4runtime_pb2 import WriteRequest, StreamMessageRequest, WriteResponse, ReadResponse, \
    SetForwardingPipelineConfigResponse, GetForwardingPipelineConfigResponse, StreamMessageResponse, \
    SetForwardingPipelineConfigRequest, GetForwardingPipelineConfigRequest, ReadRequest

from mapr import mapr_start


def fail_on_grpc_error(f):
    """
    Function annotation that fails a test in case of a gRPC runtime error
    """

    @wraps(f)
    def handle(*args, **kwargs):
        test = args[0]
        assert isinstance(test, unittest.TestCase)
        try:
            return f(*args, **kwargs)
        except grpc.RpcError as e:
            test.fail("gRPC Error: %s %s" % (e.code(), e.details()))

    return handle


class MockP4RuntimeServicer(p4runtime_pb2_grpc.P4RuntimeServicer):
    """
    Mock target P4Runtime servicer
    """

    def Write(self, request, context):
        return WriteResponse()

    def Read(self, request, context):
        for i in range(3):
            yield ReadResponse()

    def SetForwardingPipelineConfig(self, request, context):
        return SetForwardingPipelineConfigResponse()

    def GetForwardingPipelineConfig(self, request, context):
        return GetForwardingPipelineConfigResponse()

    def StreamChannel(self, request_iterator, context):
        for i in range(3):
            yield StreamMessageResponse()


class MaprTest(unittest.TestCase):
    """
    Base class for all mapr test cases
    """
    mapr = None
    stub = None
    mock_server = None

    def setUp(self):
        mapr_port = 280001
        target_port = 28000
        # stub -> 28001 (mapr) -> 28000 (mock_server)
        self.mock_server = grpc.server(futures.ThreadPoolExecutor(max_workers=1))
        p4runtime_pb2_grpc.add_P4RuntimeServicer_to_server(
            MockP4RuntimeServicer(),
            self.mock_server,
        )
        print('Starting P4RT mock server. Listening on port %d...' % target_port)
        self.mock_server.add_insecure_port('[::]:%d' % target_port)
        self.mock_server.start()

        self.mapr = mapr_start(target_addr="localhost:%d" % target_port, server_port=mapr_port)
        self.stub = p4runtime_pb2_grpc.P4RuntimeStub(
            grpc.insecure_channel("localhost:%d" % mapr_port))

    def tearDown(self):
        if self.mapr:
            self.mapr.stop(0)
        if self.mock_server:
            self.mock_server.stop(0)


class WriteTest(MaprTest):

    @fail_on_grpc_error
    def test(self):
        req = WriteRequest()
        response = self.stub.Write(req)
        self.assertIsInstance(response, WriteResponse)


class ReadTest(MaprTest):

    @fail_on_grpc_error
    def test(self):
        req = ReadRequest()
        for response in self.stub.Read(req):
            self.assertIsInstance(response, ReadResponse)


class SetForwardingPipelineConfigTest(MaprTest):

    @fail_on_grpc_error
    def test(self):
        req = SetForwardingPipelineConfigRequest()
        response = self.stub.SetForwardingPipelineConfig(req)
        self.assertIsInstance(response, SetForwardingPipelineConfigResponse)


class GetForwardingPipelineConfigTest(MaprTest):

    @fail_on_grpc_error
    def test(self):
        req = GetForwardingPipelineConfigRequest()
        response = self.stub.GetForwardingPipelineConfig(req)
        self.assertIsInstance(response, GetForwardingPipelineConfigResponse)


class StreamChannelTest(MaprTest):

    @fail_on_grpc_error
    def test(self):

        def req_iterator():
            for i in range(3):
                yield StreamMessageRequest()

        for res in self.stub.StreamChannel(req_iterator()):
            self.assertIsInstance(res, StreamMessageResponse)


if __name__ == '__main__':
    unittest.main()
