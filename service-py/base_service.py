import pb.service_pb2 as service_pb2
import pb.service_pb2_grpc as service_pb2_grpc
import grpc

class BaseService(service_pb2_grpc.ServiceServicer):
    def __init__(self):
        self.name = 'BaseService'
        self.version = 'null'
        self.concurrency = 0
        self.return_stream = False
        
    async def Call(self, 
             request: service_pb2.ServiceCallRequest, 
             context: grpc.aio.RpcContext) -> service_pb2.ServiceCallResponse:
        raise NotImplementedError("not implemented")
    
    async def CallStream(self,
                   request: service_pb2.ServiceCallRequest, 
                   context: grpc.aio.RpcContext) -> service_pb2.ServiceCallResponse:
        raise NotImplementedError("not implemented")
