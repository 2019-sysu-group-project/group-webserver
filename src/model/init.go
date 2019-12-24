package model

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

// redis 默认是没有密码和使用0号db
var Redis_client *redis.Client
var Mysql_client *sql.DB
var GormDB *gorm.DB

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
	//连接mysql
	times := 1
	for err := connectDB(); err != nil; times++ {
		if times == maxConnectionTime {
			panic(fmt.Sprint("can not connect to db after ", times, " times"))
			os.Exit(1)
			// break
		}
		log.Print("connect database with error", err, "reconnecting...")
	}
	// 将gorm调用的接口实时对应输出为真正执行的sql语句，用于debug使用
	GormDB.LogMode(true)
	fmt.Println("Starting server")
	// time.Sleep(3 * time.Second)
	times = 1
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

func connectDB() error {
	/* Database config */
	var Db_name = "projectdb"
	var Db_user = "root"
	var Db_password = "123"
	var MySqlLocation = "127.0.0.1"
	var MySqlPort = "13306"

	var Dbconnection = Db_user + ":" + Db_password + "@tcp(" + MySqlLocation + ":" + MySqlPort + ")/" + Db_name
	db, err := gorm.Open("mysql", Dbconnection+"?charset=utf8&parseTime=True") //这里的True首字母要大写！
	if err != nil {
		return err
	}
	//db.AutoMigrate(&User{}).AutoMigrate(&Product{}).AutoMigrate(&Service{})
	GormDB = db
	return nil
}
