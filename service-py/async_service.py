import asyncio
import logging
import sys

import grpc

import pb.service_pb2 as service_pb2
import pb.registry_pb2 as registry_pb2
import pb.registry_pb2_grpc as registry_pb2_grpc
from base_service import BaseService
from service_common import ServiceInstance

class AsyncService(BaseService):
    def __init__(self, registry_address):
        self.name = 'AsyncService'
        self.version = '1.0.0'
        self.concurrency = 10
        self.call_type = registry_pb2.ASYNC
        self.registry_address = registry_address
        self.init_stub()

    def init_stub(self):
        self.channel = grpc.aio.insecure_channel(self.registry_address)
        self.stub = registry_pb2_grpc.RegistryStub(self.channel)

    async def CallAsync(self, 
             request: service_pb2.ServiceCallRequest, 
             context: grpc.aio.RpcContext) -> service_pb2.ServiceCallResponse:
        logging.info(f"input(async): {request.input}")

        asyncio.create_task(self.async_task(request))

        return service_pb2.ServiceCallResponse(request_id=request.request_id, output = "success")

    async def async_task(self, request: service_pb2.ServiceCallRequest):
        result = registry_pb2.SubmitResultRequest(
            name=self.name,
            request_id=request.request_id,
            end=False,  # end a call by setting this to True
            output=request.input + " not end yet"
        )
        await self.stub.SubmitResult(result)

        result = registry_pb2.SubmitResultRequest(
            name=self.name,
            request_id=request.request_id,
            end=True,
            output=request.input + " new end"
        )
        await self.stub.SubmitResult(result)


if __name__ == '__main__':
    logging.getLogger().setLevel(logging.INFO)
    registry_address = "localhost:48001"
    if len(sys.argv) > 1:
        registry_address = sys.argv[1]

    s = ServiceInstance(registry_address, AsyncService(registry_address))
    loop = asyncio.get_event_loop()
    loop.run_until_complete(s.serve())

    # asyncio.run(s.serve()) # won't work, because asyncio.run() will create a new loop, and insecure_channel() will also create one.
