package mqueue

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

type RequestMessage struct {
	username    string
	coupon      string
	uuid        string // 表示用户发起请求的唯一id
	requestTime int64  // 用户发起请求的时间
	result      int
}

func JudgeKeyExist(uuid string) bool {
	_, ok := RequestResult[uuid]
	return ok
}

// 向消息队列发送消息
func SendMessage(username, couponName, uuid string, requestTime int64) error {
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
	request.uuid = uuid
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
func ReceiveMessage(username, couponName, uuid string, requestTime int64) (error, int) {
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

	// 先判断uuid对应的用户请求是否之前被处理过了  // 与此同时设置超时自动退出
	go func() (int, error) {
		for {
			existFlag := JudgeKeyExist(uuid)
			if existFlag == true {
				return RequestResult[uuid], nil
			}
			time := time.Now().Unix()
			if time-requestTime > 40 {
				RequestResult[uuid] = -2 //-2代表超时
				return RequestResult[uuid], nil
			}
		}
	}()

	// 负责从消费者方的amqp.Delivery读取消息
	processMsg := func(msgChan <-chan amqp.Delivery) (int, error) {
		for msg := range msgChan {
			time := time.Now().Unix()
			var request RequestMessage
			err = json.Unmarshal(msg.Body, &request)
			if err != nil {
				log.Println(err)
				return 0, err
			}
			if request.uuid != uuid {
				if time-request.requestTime > 40 {
					RequestResult[uuid] = -2
				} else {
					RequestResult[uuid] = request.result
				}
				continue
			} else {
				if time-request.requestTime > 40 {
					RequestResult[uuid] = -2
					return -2, nil
				}
				if time-request.requestTime <= 40 {
					RequestResult[uuid] = request.result
					return request.result, nil
				}
			}
		}
		return -2, nil
	}

	result, err := processMsg(msgChan)
	if err != nil {
		return err, 0
	} else if err == nil && result == -2 {
		return nil, -2
	}
	return nil, result
}
