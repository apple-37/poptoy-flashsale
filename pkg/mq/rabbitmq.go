package mq

import (
	"log"

	"poptoy-flashsale/pkg/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

var Conn *amqp.Connection
var Channel *amqp.Channel

// 队列与交换机名称定义 (严格对齐 SDD)
const (
	// 秒杀正常下单
	OrderTaskQueue    = "order.task.queue"
	
	// 延迟与死信 (15分钟超时)
	OrderDelayQueue   = "order.delay.queue"   // 消息停留在此队列，没有消费者，过期后变成死信
	OrderDlxExchange  = "order.dlx.exchange"  // 死信交换机
	OrderDlxQueue     = "order.dlx.queue"     // 死信最终进入的队列，由 Worker 消费用于取消订单
	OrderDlxRouteKey  = "order.cancel.route"
)

// InitRabbitMQ 初始化并建立拓扑结构
func InitRabbitMQ() {
	var err error
	Conn, err = amqp.Dial(config.GlobalConfig.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("RabbitMQ 连接失败: %v", err)
	}

	Channel, err = Conn.Channel()
	if err != nil {
		log.Fatalf("RabbitMQ 打开 Channel 失败: %v", err)
	}

	setupTopology()
	log.Println("RabbitMQ 连接成功, 交换机与死信队列拓扑构建完毕!")
}

// setupTopology 构建队列和绑定关系
func setupTopology() {
	// 1. 声明正常的秒杀下单队列
	_, err := Channel.QueueDeclare(OrderTaskQueue, true, false, false, false, nil)
	failOnError(err, "声明 order.task.queue 失败")

	// 2. 声明死信交换机 (Direct)
	err = Channel.ExchangeDeclare(OrderDlxExchange, "direct", true, false, false, false, nil)
	failOnError(err, "声明 order.dlx.exchange 失败")

	// 3. 声明死信队列 (接收超时的订单)
	_, err = Channel.QueueDeclare(OrderDlxQueue, true, false, false, false, nil)
	failOnError(err, "声明 order.dlx.queue 失败")

	// 4. 将死信队列绑定到死信交换机
	err = Channel.QueueBind(OrderDlxQueue, OrderDlxRouteKey, OrderDlxExchange, false, nil)
	failOnError(err, "绑定死信队列失败")

	// 5. 声明延迟队列 (配置 TTL 和 DLX)
	// 消息放入此队列后，15分钟(900000ms)后过期，过期后自动发送给 OrderDlxExchange
	args := amqp.Table{
		"x-dead-letter-exchange":    OrderDlxExchange,
		"x-dead-letter-routing-key": OrderDlxRouteKey,
		"x-message-ttl":             900000, // 15分钟 (毫秒)
	}
	_, err = Channel.QueueDeclare(OrderDelayQueue, true, false, false, false, args)
	failOnError(err, "声明 order.delay.queue 失败")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}

// Close 关闭连接
func Close() {
	if Channel != nil {
		Channel.Close()
	}
	if Conn != nil {
		Conn.Close()
	}
}