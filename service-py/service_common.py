import pb.registry_pb2_grpc as registry_pb2_grpc
import pb.registry_pb2 as registry_pb2
import grpc
import pb.service_pb2_grpc as service_pb2_grpc

from base_service import BaseService

import asyncio

class ServiceInstance(object):
    def __init__(self, registry_address: str, service_handler: BaseService):
        self.registry_address = registry_address
        self.service_handler = service_handler
        self.port = -1
        
    async def register(self):
        if self.port < 0:
            return None
        async with grpc.aio.insecure_channel(self.registry_address) as channel:
            stub = registry_pb2_grpc.RegistryStub(channel)

            service = registry_pb2.RegisterServiceRequest(
                name=self.service_handler.name, 
                port=self.port, 
                version=self.service_handler.version, 
                concurrency=self.service_handler.concurrency, 
                call_type=self.service_handler.call_type)
            await stub.RegisterService(service)

    async def register_routine(self):
        while True:
            await asyncio.sleep(5)
            await self.register()

    async def start_server(self):
        self.grpc_server = grpc.aio.server()
        service_pb2_grpc.add_ServiceServicer_to_server(self.service_handler, self.grpc_server)
        self.port = self.grpc_server.add_insecure_port("localhost:0")
        await self.grpc_server.start()
        await self.grpc_server.wait_for_termination()
        
    async def serve(self):
        await asyncio.gather(self.register_routine(), self.start_server())
        
