package model

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/go-redis/redis"
	"github.com/streadway/amqp"
)

// redis 默认是没有密码和使用0号db
var Redis_client *redis.Client
var Mysql_client *sql.DB

var MQConnection *amqp.Connection
var RequestResult = make(map[string]int)

var maxConnectionTime = 5

func init() {
	fmt.Println("init函数2被执行")
	// time.Sleep(time.Second * 5)
	// fmt.Println("Finish init server")
	Redis_client = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:16379",
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
	Mysql_client, err = sql.Open("mysql", "root:123@tcp(127.0.0.1:13306)/projectdb")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	_, err = Mysql_client.Query("SELECT * FROM User")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
}
