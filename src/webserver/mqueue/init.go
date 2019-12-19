package mqueue

import (
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/streadway/amqp"
)

var MQConnection *amqp.Connection
var RequestResult = make(map[string]int)

var maxConnectionTime = 5

func init() {
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
	conn, err := amqp.Dial("amqp://guest:guest@127.0.0.1:35672/")
	if err != nil {
		log.Println(err)
		return err
	}
	MQConnection = conn
	return nil
}
