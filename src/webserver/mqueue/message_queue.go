package mqueue

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type RequestMessage struct {
	username    string
	coupon      string
	requestTime int64 // 用户发起请求的时间
	result int
}

// 向消息队列发送消息
func SendMessage(username, couponName string, requestTime int64) error {
	//创建Channel，如果所有的只用一个channel会怎么样？
	ch, err := MQConnection.Channel()
	if err != nil {
		log.Println(err)
		return err
	}
	// 队列声明
	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		log.Println(err)
		return err
	}
	var request RequestMessage
	request.username = username
	request.coupon = couponName
	request.requestTime = requestTime
	b, err := json.Marshal(request)
	if err != nil {
		fmt.Println("error:", err)
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key  可以直接用队列名做routekey?这是默认情况吗,没有声明的时候routing key为队列名称
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        b,
		})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// 从消息队列接收消息
func ReceiveMessage(username, couponName string, requestTime int64) (error, int) {
	startTime := time.Now()
	//创建Channel，如果所有的只用一个channel会怎么样？
	ch, err := MQConnection.Channel()
	if err != nil {
		log.Println(err)
		return err, 0
	}
	// 队列声明
	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		log.Println(err)
		return err, 0
	}

	msgChan, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Println(err)
		return err, 0
	}

	// 还需要设置超时自动退出
	processMsg :=func(<-chan Delivery msgChan)int {
		for msg := range msgChan {
			var request RequestMessage
			err = json.Unmarshal(msg.Body, &request)
			if err != nil {
				log.Println(err)
				return -10
			}
			if request.username != username || request.coupon != couponName || request.requestTime != requestTime {
				continue
			} else {
				return request.result
			}
		}
		return -10
	}
	result, err := go processMsg(msgChan)
	if err != nil {
		return err, 0
	}
	return nil, result
}
