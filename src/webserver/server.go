package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"server/mqueue"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	uuid "github.com/satori/go.uuid"
)

// redis 默认是没有密码和使用0号db
var redis_client *redis.Client
var mysql_client *sql.DB

func init() {
	redis_client = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:16379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	var err error
	_, err = redis_client.Ping().Result()
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Error open redis connection")
		os.Exit(-1)
	}
	mysql_client, err = sql.Open("mysql", "root:123@tcp(127.0.0.1:13306)/projectdb")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	_, err = mysql_client.Query("SELECT * FROM User")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
}

// mysql 默认用户：root，密码：root，数据库：projectdb
// var mysql_client =

type Coupon struct {
	username    string
	coupons     string
	amount      int32
	stock       float32
	left        int32
	description string
}

// hashset 存储元组(用户名, Coupon)
var hashset map[string]string

// 任务1
func registerUser(c *gin.Context) {
	fmt.Println("This is registerUser")
}

// 任务1
func validateJWT() bool {
	// 需要编写JWT的验证机制，作为其他人能调用的一部分
	return true
}

// 任务1
func userLogin(c *gin.Context) {

}

// 任务2
func createCoupons(c *gin.Context) {

}

// 任务2
func getCouponsInformation(c *gin.Context) {

}

// 任务2
func getCouponsFromRedis(username string, coupons string) Coupon {
	return Coupon{}
}

// 任务2
func setCouponsToRedis(usernam string, coupons Coupon) {

}

// 任务2
func getCouponsFromRedisOrDatabase(username string, coupons string) Coupon {
	return Coupon{}
}

// 任务3 - 使用getCouponsFromRedis和setCouponsToRedis来完成该任务
func setCouponsToRedisAndDatabase(coupon Coupon) bool {
	// true set成功，false set失败
	return true
}

// 任务3
func patchCoupons(c *gin.Context) {
	var err error
	user, err := validateJWT()
	// 5xx: 服务端错误
	if err != nil {
		c.JSON(504, gin.H{"errMsg": "Gateway Timeout"})
		return
	}
	// TODO 401: 认证失败
	if user.author == false {
		c.JSON(401, gin.H{"errMsg": "Authorization Failed"})
		return
	}

	var coupon Coupon
	username := c.Param("username")
	name := c.Param("name")
	// 204: 已经有了优惠券
	_, exists := hashset[user] //hashset[user.username]
	if exists {
		c.JSON(204, gin.H{"errMsg": "Already had the same coupon"})
		return
	}

	coupon, err = getCouponsFromRedisOrDatabase(username, name)
	// 5xx: 服务端错误
	if err != nil {
		c.JSON(504, gin.H{"errMsg": "Gateway Timeout"})
		return
	}
	// 204: 优惠券无库存
	if coupon.left == 0 {
		c.JSON(204, gin.H{"errMsg": "The coupon is out of stock"})
		return
	}

	coupon.left--
	write, err := setCouponsToRedisAndDatabase(coupon, time.Now().UnixNano())
	// 5xx: 服务端错误
	if err != nil {
		c.JSON(504, gin.H{"errMsg": "Gateway Timeout"})
		return
	}
	// 204: 未抢到
	if write == false {
		c.JSON(204, gin.H{"errMsg": "Patch Failed"})
		return
	}

	// 将用户请求转发到消息队列中，等待消息队列对mysql进行操作并返回结果
	t := time.Now()
	// 生成uuid
	u := uuid.NewV4()
	uid := u.String()
	// 先判断是否能成功发送消息
	err = mqueue.SendMessage(username, name, uid, t.Unix())
	if err != nil {
		c.JSON(504, gin.H{"errMsg": "Gateway Timeout"})
		return
	}
	err, res := mqueue.ReceiveMessage(username, name, uid, t.Unix())
	if err != nil {
		c.JSON(504, gin.H{"errMsg": "Gateway Timeout"})
		return
	}

	//返回0代表优惠券数目为0，返回2代表抢券成功，返回1代表用户已经抢到该券不可重复抢，返回-1代表数据库访问错误，返回-2代表超时
	switch res {
	case -2:
		c.JSON(504, gin.H{"errMsg": "Mysql Server error"})
		return
	case -1:
		c.JSON(504, gin.H{"errMsg": "Mysql Server error"})
		return
	case 0:
		c.JSON(204, gin.H{"errMsg": "The coupon is out of stock"})
		return
	case 1:
		c.JSON(204, gin.H{"errMsg": "Already had the same coupon"})
		return
	case 2:
		// 201: 成功抢到
		c.JSON(201, gin.H{"errMsg": "Patch Succeeded"})
		return
	default:
		c.JSON(504, gin.H{"errMsg": "Gateway Timeout"})
		return
	}

}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.PATCH("/api/users/:username/coupons/:name", patchCoupons)
	router.POST("/api/users", registerUser)

	router.POST("/api/auth", userLogin)
	router.POST("/api/users/:username/coupons", createCoupons)

	router.GET("/api/users/:username/coupons", getCouponsInformation)
	return router
}

func main() {
	// gin.SetMode(gin.ReleaseMode)
	router := setupRouter()
	err := router.Run(":8080")
	if err != nil {
		fmt.Println("Error starting server")
	}
}
