package main

import (
	"database/sql"
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
	mysql_client, err = sql.Open("mysql", "root:123@tcp(127.0.0.1:3306)/projectdb")
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
	Username string
	Password string
	Kind     int
}

// hashset 存储元组(用户名, 商家名_优惠券名)
var hashset = make(map[string]string)

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
		kind := json.Kind
		if isUserExist(username) { // user already exists
			c.JSON(400, gin.H{
				"errMsg": "Username already exists!",
			})
		} else {
			passwordHash := md5Hash(password)

			// insert user to DB
			if insertUser(username, passwordHash, kind) {
				c.JSON(201, gin.H{
					"errMsg": "",
				})
			} else {
				c.JSON(400, gin.H{
					"errMsg": "Create user failed!",
				})
			}
		}
	}
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
		if authenticateUser(json.Username, json.Password) {
			token := genToken()
			if token == "" {
				c.JSON(401, gin.H{
					"kind":   "",
					"errMsg": "Generate token failed.",
				})
			} else {
				var user User
				err := mysql_client.QueryRow("SELECT kind FROM User WHERE username=?", json.Username).Scan(&user.Kind)
				if err != nil {
					c.JSON(500, gin.H{
						"kind":   "",
						"errMsg": "Query DB failed.",
					})
				} else {
					c.JSON(200, gin.H{
						"Authorization": token,
						"kind":          user.Kind,
						"errMsg":        "",
					})
				}
			}
		} else {
			c.JSON(401, gin.H{
				"kind":   "",
				"errMsg": "Login failed.",
			})
		}
	}
}

// check if the user already exists in DB
func isUserExist(query_username string) bool {
	var user User
	err := mysql_client.QueryRow("SELECT username, password, kind FROM User WHERE username=?", query_username).Scan(&user.Username, &user.Password, &user.Kind)
	if err == sql.ErrNoRows { // user not exists
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
		if user.Username == username {
			if user.Password == passwordHash {
				return true
			}
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
func setCouponsToRedisAndDatabase(coupon Coupon, time int64) bool {
	// true set成功，false set失败

	return true
}

// 任务3
func patchCoupons(c *gin.Context) {
	var err error
	// 5xx: 服务端错误
	if err != nil {
		c.JSON(504, gin.H{"errMsg": "Gateway Timeout"})
		return
	}
	// 401: 认证失败
	if validateJWT(c) == false {
		c.JSON(401, gin.H{"errMsg": "Authorization Failed"})
		return
	}

	username := c.Param("username")
	name := c.Param("name")
	// TODO: user := xxx(c)
	// 204: 已经有了优惠券
	_, exists := hashset[user.username] //hashset[user.username]
	if exists {
		c.JSON(204, gin.H{"errMsg": "Already had the same coupon"})
		return
	}

	var coupon = getCouponsFromRedisOrDatabase(username, name)
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
	write := setCouponsToRedisAndDatabase(coupon, time.Now().UnixNano())
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
	// 201: 成功抢到
	if write == true {
		hashset[user.username] = username + "_" + name
		c.JSON(201, gin.H{"errMsg": "Patch Succeeded"})
		return
	}
}

func printHello(c *gin.Context) {
	c.String(200, "hello")
}

func setupRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/", printHello)

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
