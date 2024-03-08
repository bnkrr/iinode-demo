import asyncio
import logging
import typing
import sys

import grpc

import pb.service_pb2 as service_pb2
from base_service import BaseService
from service_common import ServiceInstance


class StreamService(BaseService):
    def __init__(self):
        self.name = 'StreamService'
        self.version = '1.0.0'
        self.concurrency = 1
        self.return_stream = True
        
    async def CallStream(self, 
            request: service_pb2.ServiceCallRequest, 
            context: grpc.aio.RpcContext) -> typing.AsyncGenerator[service_pb2.ServiceCallResponse,None]:
        logging.info(f"stream request: {request.input}")
        for i in range(3):
            out_str = f'result {i} for input {request.input}'
            yield service_pb2.ServiceCallResponse(output = out_str)

if __name__ == '__main__':
    logging.getLogger().setLevel(logging.INFO)
    registry_address = "localhost:48001"
    if len(sys.argv) > 1:
        registry_address = sys.argv[1]

    s = ServiceInstance(registry_address, StreamService())
    asyncio.run(s.serve())


