package main

import (
	"database/sql"
	"fmt"
	"os"

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

//2
//将Coupon拼接成字符串，以#分隔：...#...#...
func (c *Coupon) ToString() string {
	var s string
	s = fmt.Sprintf("%s#%s#%d#%f#%d#%s", c.username, c.coupons, c.amount,
		c.stock, c.left, c.description)
	return s
}
//2
//将字符串转换成Coupon
func (c *Coupon) ToCoupon(s string) {
	j := strings.LastIndex(s, "#")
	c.description = s[j+1:]
	s = s[:j]
	j = strings.LastIndex(s, "#")
	left, _ := strconv.ParseInt(s[j+1:], 10, 32)
	c.left = int32(left)
	s = s[:j]
	j = strings.LastIndex(s, "#")
	stock, _ := strconv.ParseFloat(s[j+1:], 32)
	c.stock = float32(stock)
	s = s[:j]
	j = strings.LastIndex(s, "#")
	amount, _ := strconv.ParseInt(s[j+1:], 10, 32)
	c.amount = int32(amount)
	s = s[:j]
	j = strings.LastIndex(s, "#")
	c.coupons = s[j+1:]
	s = s[:j]
	c.username = s
	fmt.Println(c.description)
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
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

// 任务2
func getCouponsInformation(c *gin.Context) {
	var couponJSON Coupon
	couponJSON.username=c.Param("username")
	couponJSON.coupons=c.Param("coupons")
	coupon := getCouponsFromRedisOrDatabase(couponJSON.username, couponJSON.coupons)
	c.JSON(http.StatusOK, gin.H{"username": coupon.username, "coupons": coupon.coupons,
		"amount": coupon.amount, "stock": coupon.stock, "left_coupons": coupon.left,
		"description": coupon.description})
}

// 任务2
func getCouponsFromRedis(username string, cou string) (Coupon, error) {
	re, err := redis_client.Get(username + " " + cou).Result()
	var result Coupon
	result.ToCoupon(re)
	return result, err
}

// 任务2
func setCouponsToRedis(username string, cou Coupon) {
	redis_client.Set(username+" "+cou.coupons, cou.ToString(), 1000)
}

// 任务2
func getCouponsFromRedisOrDatabase(username string, cou string) Coupon {
	var result Coupon
	result, err := getCouponsFromRedis(username, cou)
	if err != nil {
		query, _ := mysql_client.Query("SELECT * FROM Coupon WHERE username=? AND coupons=?", username, cou)
		query.Scan(&result.username, &result.coupons, &result.amount,
			&result.stock, &result.left, &result.description)
		setCouponsToRedis(username, result)
	}
	return result
}

// 任务3 - 使用getCouponsFromRedis和setCouponsToRedis来完成该任务
func setCouponsToRedisAndDatabase(coupon Coupon) bool {
	// true set成功，false set失败
	return true
}

// 任务3
func patchCoupons(c *gin.Context) {

}

func setupRouter() *gin.Engine{
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
