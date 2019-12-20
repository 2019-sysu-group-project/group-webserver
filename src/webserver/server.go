package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"

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
	// mysql_client, err = sql.Open("mysql", "root:123@tcp(127.0.0.1:13306)/projectdb")
	// test
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
	Username string `json:"username"`
	Password string `json:"password"`
	Kind     string `json:"kind"`
}

type User_DB struct {
	Username string
	Password string
	Kind     int
}

var (
	key []byte = []byte("JWT key of GXNXFD")
)

// 任务1
func registerUser(c *gin.Context) {
	// fmt.Println("This is registerUser")
	var json User

	if c.BindJSON(&json) == nil {
		username := json.Username
		password := json.Password
		kind := json.Kind // string type
		if username == "" || password == "" {
			c.JSON(400, gin.H{
				"errMsg": "空用户名或密码",
			})
			return
		}
		var kindInt int
		if kind == "customer" || kind == "" {
			kindInt = 0
		} else if kind == "saler" {
			kindInt = 1
		} else { // wrong type of kind
			c.JSON(400, gin.H{
				"errMsg": "错误kind类型",
			})
			return
		}
		if isUserExist(username) { // user already exists
			c.JSON(400, gin.H{
				"errMsg": "用户已存在",
			})
			return
		}
		passwordHash := md5Hash(password)
		// insert user to DB
		if insertUser(username, passwordHash, kindInt) {
			c.JSON(201, gin.H{
				"errMsg": "",
			})
			return
		}
		c.JSON(400, gin.H{
			"errMsg": "创建账户失败",
		})
		return
	}
	// failed in BindJSON
	c.JSON(400, gin.H{
		"errMsg": "获取json数据失败",
	})
	return
}

// generate JWT token
func genToken() string {
	claims := &jwt.StandardClaims{
		NotBefore: int64(time.Now().Unix()),
		ExpiresAt: int64(time.Now().Unix() + 3600),
		Issuer:    "GXNXFD",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(key)
	if err != nil {
		return ""
	}
	return ss
}

// 任务1
func validateJWT(c *gin.Context) bool {
	// 需要编写JWT的验证机制，作为其他人能调用的一部分
	token := c.Request.Header.Get("Authorization")
	_, err := jwt.Parse(token, func(*jwt.Token) (interface{}, error) {
		return key, nil
	})
	if err != nil {
		return false
	}
	return true
}

// 任务1
func userLogin(c *gin.Context) {
	var json User
	if c.BindJSON(&json) == nil {
		if json.Username == "" || json.Password == "" {
			c.JSON(401, gin.H{
				"kind":   "",
				"errMsg": "空用户或空密码",
			})
			return
		}
		if isUserExist(json.Username) == false {
			c.JSON(401, gin.H{
				"kind":   "",
				"errMsg": "用户不存在",
			})
			return
		}
		if authenticateUser(json.Username, json.Password) {
			token := genToken()
			if token == "" {
				c.JSON(401, gin.H{
					"kind":   "",
					"errMsg": "Generate token failed.",
				})
				return
			} else {
				var user User_DB
				err := mysql_client.QueryRow("SELECT kind FROM User WHERE username=?", json.Username).Scan(&user.Kind)
				if err != nil {
					c.JSON(500, gin.H{
						"kind":   "",
						"errMsg": "Query DB failed.",
					})
					return
				} else {
					var kindString string
					if user.Kind == 0 {
						kindString = "customer"
					} else if user.Kind == 1 {
						kindString = "saler"
					} else {
						c.JSON(500, gin.H{
							"kind":   "",
							"errMsg": "Wrong kind in DB.",
						})
					}
					c.Header("Authorization", token)
					c.JSON(200, gin.H{
						// "Authorization": token,
						"kind":   kindString,
						"errMsg": "",
					})
					return
				}
			}
		} else {
			c.JSON(401, gin.H{
				"kind":   "",
				"errMsg": "错误密码",
			})
			return
		}
	} else { // failed in BindJSON
		c.JSON(400, gin.H{
			"errMsg": "获取json数据失败",
		})
		return
	}
}

// check if the user already exists in DB
func isUserExist(usernameQuery string) bool {
	var user User_DB
	err := mysql_client.QueryRow("SELECT username, password, kind FROM User WHERE username=?", usernameQuery).Scan(&user.Username, &user.Password, &user.Kind)
	if err != nil {
		fmt.Println(err)
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

// insert user into DB
func insertUser(username string, password string, kind int) bool {
	result, err := mysql_client.Exec("INSERT INTO User(username, password, kind) VALUES(?,?,?)", username, password, kind)
	if err != nil {
		// insert failed
		return false
	}
	_, err = result.LastInsertId()
	if err != nil {
		return false
	}
	_, err = result.RowsAffected()
	if err != nil {
		return false
	}
	return true
}

// authenticate user from DB
func authenticateUser(username string, password string) bool {
	passwordHash := md5Hash(password)
	var user User
	err := mysql_client.QueryRow("SELECT username, password FROM User WHERE username=?", username).Scan(&user.Username, &user.Password)
	if err == sql.ErrNoRows {
		return false
	} else {
		if user.Username == username && user.Password == passwordHash {
			return true
		}
	}
	return false
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
}
