# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: pb/service.proto
# Protobuf Python Version: 4.25.0
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x10pb/service.proto\x12\npb_autogen\"7\n\x12ServiceCallRequest\x12\r\n\x05input\x18\x01 \x01(\t\x12\x12\n\nrequest_id\x18\x02 \x01(\x05\"9\n\x13ServiceCallResponse\x12\x0e\n\x06output\x18\x01 \x01(\t\x12\x12\n\nrequest_id\x18\x02 \x01(\x05\x32\xf7\x01\n\x07Service\x12I\n\x04\x43\x61ll\x12\x1e.pb_autogen.ServiceCallRequest\x1a\x1f.pb_autogen.ServiceCallResponse\"\x00\x12Q\n\nCallStream\x12\x1e.pb_autogen.ServiceCallRequest\x1a\x1f.pb_autogen.ServiceCallResponse\"\x00\x30\x01\x12N\n\tCallAsync\x12\x1e.pb_autogen.ServiceCallRequest\x1a\x1f.pb_autogen.ServiceCallResponse\"\x00\x42)Z\'github.com/bnkrr/iinode-demo/pb_autogenb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'pb.service_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:
  _globals['DESCRIPTOR']._options = None
  _globals['DESCRIPTOR']._serialized_options = b'Z\'github.com/bnkrr/iinode-demo/pb_autogen'
  _globals['_SERVICECALLREQUEST']._serialized_start=32
  _globals['_SERVICECALLREQUEST']._serialized_end=87
  _globals['_SERVICECALLRESPONSE']._serialized_start=89
  _globals['_SERVICECALLRESPONSE']._serialized_end=146
  _globals['_SERVICE']._serialized_start=149
  _globals['_SERVICE']._serialized_end=396
# @@protoc_insertion_point(module_scope)
