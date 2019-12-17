package main

import (
	"os"
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"

	"net/http"
	"crypto/md5"
	"io"
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

type User struct {
	username	string
	password	string
	kind		int
}

// 任务1
func registerUser(c *gin.Context) {
	// fmt.Println("This is registerUser")
	var json User

	if c.BindJSON(&json) == nil {
		username := json.username
		password := json.password
		kind := json.kind
		if isUserExist(username) {		// user already exists
			c.JSON(400, gin.H{
				"errMsg": "Username already exists!",
			})
		} else {
			passwordHash := md5Hash(password)

			// insert user to DB
			if insertUser(username, passwordHash, kind) {
				c.JSON(200, gin.H{
					"errMsg": "",
				})
			} else {
				c.JSON(400, gin.H{
					"errMsg": "Create user failed!"
				})
			}
		}
	}
}

// 任务1
func validateJWT(username string, password string) bool {
	// 需要编写JWT的验证机制，作为其他人能调用的一部分
	return true
}

// 任务1
func userLogin(c *gin.Context) {
	var json User
	if c.BindJSON(&json) == nil {
		if validateJWT(json.username, json.password) {
			c.JSON(200, gin.H{
				"kind": "",
				"errMsg": "",
			})
		} else {
			c.JSON(401, gin.H{
				"kind": "",
				"errMsg": "Login failed.",
			})
		}
	}
}

// check if the user already exists in DB
func isUserExist(query_username string) bool {
	// undo: query user from DB
	var user User

	err := mysql_client.QueryRow("SELECT username, password, kind FROM User WHERE username=?", query_username).Scan(&user.username, &user.password, &user.kind)

	if err == sql.ErrNoRows {		// user not exists
		return false
	} else {
		return true
	}
}

// md5
func md5Hash(data string) string {
	hash := md5.New()
	io.WriteString(hash, data)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func insertUser(username string, password string, kind int) {
	// insert user to DB
	result, err := mysql_client.Exec("INSERT INTO User(username, password, kind) VALUES(?,?,?)", username, password, kind)
	if err != nil {
		// insert failed
		return false
	}
	lastInsertID, err := result.lastInsertId()
	if err != nil {
		return false
	}
	rowsaffected, err := result.RowsAffected()
	if err != nil {
		return false
	}
	return true
}

// 任务2
func createCoupons(c *gin.Context) {

}

// 任务2
func getCouponsInformation(c *gin.Context) {

}

// 任务2
func getCouponsFromRedisOrDatabase(username string, coupons string) Coupon {
	return Coupon{}
}

// 任务3
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