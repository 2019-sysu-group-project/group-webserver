package model

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/streadway/amqp"
)

// redis 默认是没有密码和使用0号db
var Redis_client *redis.Client
var Mysql_client *sql.DB

var MQConnection *amqp.Connection
var RequestResult = make(map[string]int)

var maxConnectionTime = 5

func init() {
	time.Sleep(30 * time.Second)
	Redis_client = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	var err error
	_, err = Redis_client.Ping().Result()
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Error open redis connection")
		os.Exit(-1)
	}
	//mysql_client, err = sql.Open("mysql", "root:123@tcp(projectdb:3306)/projectdb")
	Mysql_client, err = sql.Open("mysql", "root:123@tcp(db:3306)/projectdb")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	_, err = Mysql_client.Query("SELECT * FROM User")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	fmt.Println("Starting server")
	// fmt.Println("init函数3被执行")
	// time.Sleep(3 * time.Second)
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
