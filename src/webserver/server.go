package main

import (
	"os"
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
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

//实现Reader接口，用于redis_client.Set("", Coupon,10)
type (c * Coupon) Read(p []byte)(n int, err error){
    d:=[]byte(*c)
    var i int
    for i, v := range d{
        p[i]=v
    }
    return i,nil
}

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
	var couponJSON Coupon
	err := c.ShouldBind(&couponJSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		mysql_client.Query("insert into Coupon values(?,?,?,?,?,?)",
			couponJSON.username, couponJSON.coupons, couponJSON.amount,
			couponJSON.stock, couponJSON.left, couponJSON.description)
	}
}

// 任务2

func getCouponsInformation(c *gin.Context) {
	var couponJSON Coupon
	err := c.ShouldBind(&couponJSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		coupon := getCouponsFromRedisOrDatabase(couponJSON.username, couponJSON.coupons)
		c.JSON(http.StatusOK, gin.H{"username": coupon.username, "coupons": coupon.coupons,
			"amount": coupon.amount, "stock": coupon.stock, "left_coupons": coupon.left,
			"description": coupon.description})
	}

}

// 任务2

func getCouponsFromRedisOrDatabase(username string, coupons string) Coupon {
	var result Coupon
	result, err := redis_client.Get(username + " " + coupons).Result()
	if err != nil {
		query, _ := mysql_client.Query("SELECT * FROM Coupon WHERE username=? AND coupons=?", username, coupons)
		query.Scan(&result.username, &result.coupons, &result.amount,
			&result.stock, &result.left, &result.description)
		redis_client.Set(username+" "+coupons, &result, 10000)
	}
	return result

}

// 任务3
func setCouponsToRedisAndDatabase(coupon Coupon) bool {
	// true set成功，false set失败
	return true
}

// 任务3
func patchCoupons(c *gin.Context) {

}

func main() {
	// gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.PATCH("/api/users/:username/coupons/:name", patchCoupons)
	router.POST("/api/users", registerUser)

	router.POST("/api/auth", userLogin)
	router.POST("/api/users/:username/coupons", createCoupons)

	router.GET("/api/users/:username/coupons", getCouponsInformation)

	err := router.Run(":8080")
	if err != nil {
		fmt.Println("Error starting server")
	}
}
