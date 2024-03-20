# python 版测量工具 demo

依靠async进行并发，性能会有限制。

## 环境配置

需要python3。程序在python3.10下测试通过。python需安装grpc、protoc等依赖

```
pip install grpcio-tools
```

此后可需生成protobuf代码(可参照`Makefile`)，在本目录下（`service-py/`）运行：
```
python -m grpc_tools.protoc -Ipb=../proto --python_out=. \
		--pyi_out=. --grpc_python_out=. ../proto/*.proto
```

## walkthrough

1. 运行golang版本的registry
2. 更改`echo_service.py`的代码，实现`Call()`函数。
3. 运行`python echo_service.py localhost:xxxx`，其中`xxxx`是registry端口号。

流的实现类似，可参考`stream_service.py`