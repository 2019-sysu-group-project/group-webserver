package model

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

// 从消息队列接收消息
func ReceiveMessage(username, couponName, uuid string, requestTime int64) (error, int) {
	//创建Channel，如果所有的只用一个channel会怎么样？
	ch, err := MQConnection.Channel()
	if err != nil {
		log.Println(err)
		return err, 0
	}
	defer ch.Close()
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

	var existFlag bool = false
	// var timeOut bool = false
	// 先判断uuid对应的用户请求是否之前被处理过了  // 与此同时设置超时自动退出
	/*go func() {
		for {
			existFlag = JudgeKeyExist(uuid)
			if existFlag == true {
				break
			}
			time := time.Now().Unix()
			if time-requestTime > 40 {
				RequestResult[uuid] = -2 //-2代表超时
				break
			}
		}
	}()*/
	// 目前消息从消息队列丢失从而导致超时即全局超时问题还没解决

	// 负责从消费者方的amqp.Delivery读取消息
	processMsg := func(msgChan <-chan amqp.Delivery) (int, error) {
		var count int
		for msg := range msgChan {
			// 计数器用于debug
			count++
			fmt.Printf("循环次数%v\n", count)
			// 判断请求之前是否被处理过了
			existFlag = JudgeKeyExist(uuid)
			if existFlag {
				return RequestResult[uuid], nil
			}

			time := time.Now().Unix()
			if time-requestTime > 40 {
				RequestResult[uuid] = -2 //-2代表超时
				return RequestResult[uuid], nil
			}
			var request RequestMessage
			err = json.Unmarshal(msg.Body, &request)
			if err != nil {
				log.Println(err)
				return 0, err
			}
			if request.Uuid != uuid {
				if time-request.RequestTime > 40 {
					RequestResult[uuid] = -2
				} else {
					RequestResult[uuid] = request.Result
				}
				continue
			} else {
				fmt.Println(33333)
				if time-request.RequestTime > 40 {
					RequestResult[uuid] = -2
					return -2, nil
				}
				if time-request.RequestTime <= 40 {
					RequestResult[uuid] = request.Result
					return request.Result, nil
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

// 向消息队列发送消息
func SendMessage(username, couponName, uuid string, requestTime int64) error {
	//创建Channel，如果所有的只用一个channel会怎么样？
	ch, err := MQConnection.Channel()
	if err != nil {
		log.Println(err)
		return err
	}
	defer ch.Close()
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
	request.Username = username
	request.Coupon = couponName
	request.Uuid = uuid
	request.RequestTime = requestTime
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

func JudgeKeyExist(uuid string) bool {
	_, ok := RequestResult[uuid]
	return ok
}
