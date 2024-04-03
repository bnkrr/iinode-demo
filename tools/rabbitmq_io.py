import pika

########### 基本信息 ##############
HOST = '1.2.3.4'      # rabbitmq服务器地址
PORT = 5672           # rabbitmq服务器端口
USER = 'user'         # rabbitmq用户名
PASSWD = 'password'   # rabbitmq密码
VHOST = 'vhost-name'  # rabbitmq vhost

PUBLISH_EXCHANGE = ''                      # 发布信息的exchange名称
PUBLISH_ROUTING_KEY = 'output_queue_name'  # 发布信息的routing key名称，如果exchange为空，则为队列名称

COMSUME_QUEUE = 'input_queue_name'         # 消费信息的队列名称


########### 初始化连接、认证 ##########
credentials = pika.PlainCredentials(USER, PASSWD)
parameters = pika.ConnectionParameters(HOST,PORT,VHOST,credentials)
connection = pika.BlockingConnection(parameters)
channel = connection.channel()


############ 向队列发布一条消息 ##############
# if PUBLISH_EXCHANGE:
#     channel.exchange_declare(PUBLISH_EXCHANGE, passive=True)
# else:
#     channel.queue_declare(PUBLISH_ROUTING_KEY, passive=True)
## 发送消息
# channel.basic_publish(exchange=PUBLISH_EXCHANGE, routing_key=PUBLISH_ROUTING_KEY, body='Hello World!')

# connection.close()  # 关闭连接


########## 从队列中消费一个信息 ##########
channel.queue_declare(COMSUME_QUEUE, passive=True)

method_frame, header_frame, body = channel.basic_get(queue=COMSUME_QUEUE, auto_ack=True)
if method_frame:
    print(f" [x] Received {body}")
else:
    print(f" [x] Nothing to do")
    
connection.close()  # 关闭连接

######### 从队列中持续消费信息 #############
# def callback(ch, method, properties, body):
#     print(f" [x] Received {body}")

# channel.queue_declare(COMSUME_QUEUE, passive=True)
# channel.basic_consume(queue=COMSUME_QUEUE, on_message_callback=callback, auto_ack=True)
# channel.start_consuming() # 开始消费

