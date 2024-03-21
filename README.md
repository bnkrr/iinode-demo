# 测量工具demo

## 总体情况

本demo给出了一个最小化的golang示例，展示测量工具如何编写。

* `service` 文件夹下是测量工具demo，更改以实现不同测量功能。
* `service-py` 文件夹下是测量工具的python版demo，更改以实现不同测量功能。
* `registry` 文件夹下是注册中心demo，支持服务注册，支持读取队列输入并调用已注册的服务，支持文件做输入输出。调试用。
* `proto` 文件夹下是protobuf数据结构，无需更改。
* `pb_autogen` 文件夹下是protobuf相关生成go代码，无需更改。

## quickstart
环境可参考网络中关于golang/grpc的内容。

更改`service/`下的文件，以实现不同的测量功能。熟悉gRPC接口后，service可以整体使用其他语言（如python、java等）编写。

运行registry：
```shell
go run registry/*.go --addr localhost:56789 --config registry/config_example.json
```

同时运行service：
```shell
go run service/*.go --reg-addr localhost:56789
```

此时可看到service向registry注册，并被registry周期性调用。

## 高级

### protobuf代码生成

* `make pb_gen` 生成protobuf相关go代码
* `make reg` 生成registry二进制文件
* `make srv` 生成service二进制文件

细节可查看`Makefile`

### registry 配置
可按照`registry/config_example.json`编写配置，以使得registry可以接入队列，根据队列输入输出调用service，或者接入文件输入输出调用service。

使用`--config config.json`来运行registry，如下：
```shell
go run registry/*.go --addr localhost:56789 --config registry/config_example.json
```

* 通过`io-type`字段指定输入输出类型
* 文件作输入输出（供调试用）
  - 通过`file-***-path`指定文件路径，每一行一个输入。registry会读文件，并根据设置多次发送每一行。
  - 通过`file-restart-***`控制反复发送的行为
* 队列作输入输出（可供测量使用）
  - 通过`mq-url`指定RabbitMQ服务器的配置。
  - 通过`mq-***`指定输入、输出队列信息

```json
{
    "services": [
        {
            "name": "EchoService",  //服务名称
            "io-type": "file", // 文件类型的输入输出，会读文件作为输入（可以重复多次），并将输出写到文件里
            "file-input-path": "path/to/input_example.txt",　// 输入文件路径
            "file-output-path": "path/to/output_example.txt", // 输出文件路径，为空则只会打印在stdout
            "file-restart-count": 0, // 重复次数，为0则无限重复
            "file-restart-interval": 5 // 重复的间隔，单位为秒
        },
        {
            "name": "StreamService", //服务名称
            "io-type": "mq",  // MQ 队列类型的输入输出，会消费队列作为输入，并将输出写到队列里
            "mq-url": "amqp://user:password@localhost:5672/vhost", // MQ队列信息
            "mq-input-queue": "input-test", // 输入队列名称
            "mq-output-exchange": "ex-test", // 输出exchange名称
            "mq-output-routing-key": "output" // 输出routing key名称（如exchange为空，则为输出队列名称）
        }
    ]
}

```

## Q&A

Q: 写一个service本质上需要做什么？
A: 实现gRPC的`Call/CallStream()`接口，并且调用`RegisterService()`接口进行注册。

Q: 最简单的运行demo的方法是什么？
A: 更改`service/echoservice.go`中`Call()`的实现。

Q: 我能写Python吗？
A: 什么语言都可以，只要能实现gRPC的接口即可。Python版demo可以见目录`service-py/`。

## Changelog
```
20240321
增加registry文件输入输出功能

20240320
增加registry调用队列功能，如不需要可暂不更新
```

## TODO
* 反向回传功能支持
* 队列功能增强
  - ACK支持
  - prefetch
