package mqueue

import (
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/streadway/amqp"
)

var MQConnection *amqp.Connection
var RequestResult = make(map[string]int)

var maxConnectionTime = 5

func init() {
	fmt.Println("Starting server")
	time.Sleep(30 * time.Second)
	times := 1
	for err := connectMQ(); err != nil; times++ {
		if times == maxConnectionTime {
			panic(fmt.Sprint("can not connect to mq after ", times, " times"))
			os.Exit(1)
			// break
		}
		log.Print("connect mq with error", err, "reconnecting...")
	}

}

func connectMQ() error {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Println(err)
		return err
	}
	MQConnection = conn
	return nil
}
