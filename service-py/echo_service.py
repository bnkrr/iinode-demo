import asyncio
import logging
import sys

import grpc

import pb.service_pb2 as service_pb2
import pb.registry_pb2 as registry_pb2
from base_service import BaseService
from service_common import ServiceInstance

class EchoService(BaseService):
    def __init__(self):
        self.name = 'EchoService'
        self.version = '1.0.0'
        self.concurrency = 1
        self.call_type = registry_pb2.NORMAL
        
    async def Call(self, 
            request: service_pb2.ServiceCallRequest, 
            context: grpc.aio.RpcContext) -> service_pb2.ServiceCallResponse:
        logging.info(f"echo request: {request.input}")
        return service_pb2.ServiceCallResponse(output = request.input)

if __name__ == '__main__':
    logging.getLogger().setLevel(logging.INFO)
    registry_address = "localhost:48001"
    if len(sys.argv) > 1:
        registry_address = sys.argv[1]

    s = ServiceInstance(registry_address, EchoService())
    asyncio.run(s.serve())


