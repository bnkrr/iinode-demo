# 测量工具demo

## 总体情况

本demo给出了一个最小化的golang示例，展示测量工具如何编写。

* `service` 文件夹下是测量工具demo，更改以实现不同测量功能。
* `service-py` 文件夹下是测量工具的python版demo，更改以实现不同测量功能。
* `registry` 文件夹下是注册中心demo，支持服务注册，会持续调用已注册的服务。调试用。
* `proto` 文件夹下是protobuf数据结构，无需更改。
* `pb_autogen` 文件夹下是protobuf相关生成go代码，无需更改。

## quickstart
环境可参考网络中关于golang/grpc的内容。

更改`service/`下的文件，以实现不同的测量功能。熟悉gRPC接口后，service可以整体使用其他语言（如python、java等）编写。

运行registry：
```shell
go run registry/*.go --addr localhost:56789
```

同时运行service：
```shell
go run service/*.go --reg-addr localhost:56789
```

此时可看到service向registry注册，并被registry周期性调用。

## 高级

* `make pb_gen` 生成protobuf相关go代码
* `make reg` 生成registry二进制文件
* `make srv` 生成service二进制文件

细节可查看`Makefile`

## Q&A

Q: 写一个service本质上需要做什么？
A: 实现gRPC的`Call/CallStream()`接口，并且调用`RegisterService()`接口进行注册。

Q: 最简单的运行demo的方法是什么？
A: 更改`service/echoservice.go`中`Call()`的实现。

Q: 我能写Python吗？
A: 什么语言都可以，只要能实现gRPC的接口即可。Python版demo可以见目录`service-py/`。
