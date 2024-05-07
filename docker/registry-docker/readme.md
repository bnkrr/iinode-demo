# registry runtime docker
* 使用`build-registry.sh`进行构建生成二进制
* 更改`config_mq1.json`和`docker-compose.yml`中的相关配置，如rabbitmq地址、本地监听地址等。

```shell
docker compose build # 构建docker image
docker compose up    # 启动（并构建）docker container
docker compose down  # 停止container
docker compose down --rmi local --remove-orphans # 停止container并删除多余的image
```