package main

import (
	"database/sql"
	"encoding/json"
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
	Username    string
	Coupons     string
	Amount      int32
	Stock       float32
	Left        int32
	Description string
}

type User struct {
	Username string
	Password string
	Kind     string
}

type User_DB struct {
	Username string
	Password string
	Kind     int
}

// 任务1
func registerUser(c *gin.Context) {

}

// 任务1
func validateJWT(c *gin.Context) bool {
	// 需要编写JWT的验证机制，作为其他人能调用的一部分
	return true
}

// 任务1
func userLogin(c *gin.Context) {

}

// check if the user already exists in DB
func isUserExist(query_username string) bool {
	var user User_DB
	err := mysql_client.QueryRow("SELECT username, password, kind FROM User WHERE username=?", query_username).Scan(&user.Username, &user.Password, &user.Kind)
	if err == sql.ErrNoRows { // user not exists
		return false
	} else {
		return true
	}
}

// 任务2
func createCoupons(c *gin.Context) {

}

// 任务2
func getCouponsInformation(c *gin.Context) {

}

// 任务2
func getCouponsFromRedis(Username string, cou string) (Coupon, error) {
	return Coupon{}, nil
}

// 任务2
func setCouponsToRedis(Username string, cou Coupon) {
}

// 任务2
func getCouponsFromRedisOrDatabase(Username string, cou string) (Coupon, error) {
	return Coupon{}, nil
}

// 任务3
func patchCoupons(c *gin.Context) {

}

func setupRouter() *gin.Engine {
	router := gin.Default()
	// task1
	router.POST("/api/users", registerUser)
	router.POST("/api/auth", userLogin)
	// task2
	router.POST("/api/users/:username/coupons", createCoupons)
	router.GET("/api/users/:username/coupons", getCouponsInformation)
	// task3
	router.PATCH("/api/users/:username/coupons/:name", patchCoupons)

	// used for testing
	router.GET("/validate", testValidateJWT)
	router.GET("/test", testMyFunc)
	return router
}

func testMyFunc(c *gin.Context) {
	s := `{"errMsg": "", "data": [{"name": "test_coupons_xxx", "amount": 100, "left": 30, "stock": 500, "description": "no description"}]}`
	var result map[string]interface{}
	err := json.Unmarshal([]byte(s), &result)
	if err != nil {
		fmt.Println(err.Error())
	}
	c.JSON(200, result)
}

func testValidateJWT(c *gin.Context) {
	valid := validateJWT(c)
	if valid {
		c.JSON(200, gin.H{
			"errMsg": "valid",
		})
	} else {
		c.JSON(200, gin.H{
			"errMsg": "invalid",
		})
	}
}

func main() {
	// gin.SetMode(gin.ReleaseMode)
	router := setupRouter()
	err := router.Run(":8080")
	if err != nil {
		fmt.Println("Error starting server")
	}
	fmt.Println("HAHAHA")
}
