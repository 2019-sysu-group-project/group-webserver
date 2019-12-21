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
	"server/mqueue"
	"strconv"

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
// var mysql_client =sql.Open("mysql", "root:123@tcp(127.0.0.1:13306)/projectdb")

// Coupon has username as shopper's name, coupons as its name
type Coupon struct {
	username    string
	coupons     string
	amount      int32
	stock       float32
	left        int32
	description string
}

// User records user's name, password and kind to idetify if it's a seller or a buyer
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Kind     string `json:"kind"`
}

// MyClaims: Customer Claims
type MyClaims struct {
	Uname string `json:"username"`
	jwt.StandardClaims
}

// User_DB not known yet
type User_DB struct {
	Username string
	Password string
	Kind     int
}

var (
	key []byte = []byte("JWT key of GXNXFD")
)

// hashset 存储元组(用户名, 商家名_优惠券名)
var hashset map[string]string

// 任务1
func registerUser(c *gin.Context) {
	// fmt.Println("This is registerUser")
	var json User

	if c.BindJSON(&json) == nil {
		username := json.Username
		password := json.Password
		kind := json.Kind          // string type
		if isUserExist(username) { // user already exists
			c.JSON(400, gin.H{
				"errMsg": "Username already exists!",
			})
			return
		} else {
			int_kind, err := strconv.Atoi(kind)
			if err != nil {
				c.JSON(400, gin.H{
					"errMsg": "Post error.",
				})
				return
			}
			passwordHash := md5Hash(password)
			// insert user to DB
			if insertUser(username, passwordHash, int_kind) {
				c.JSON(201, gin.H{
					"errMsg": "",
				})
				return
			} else {
				c.JSON(400, gin.H{
					"errMsg": "Create user failed!",
				})
				return
			}
		}
	} else { // failed in BindJSON
		c.JSON(400, gin.H{
			"errMsg": "Failed in BindJSON!",
		})
		return
	}
}

// generate JWT token
func genToken(username string) string {
	claims := &MyClaims{
		username,
		jwt.StandardClaims{
			NotBefore: int64(time.Now().Unix()),
			ExpiresAt: int64(time.Now().Unix() + 3600),
			Issuer:    "GXNXFD",
		},
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
			token := genToken(json.Username)
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
					string_kind := strconv.Itoa(user.Kind)
					c.JSON(200, gin.H{
						"Authorization": token,
						"kind":          string_kind,
						"errMsg":        "",
					})
					return
				}
			}
		} else {
			c.JSON(401, gin.H{
				"kind":   "",
				"errMsg": "Login failed.",
			})
			return
		}
	} else { // failed in BindJSON
		c.JSON(400, gin.H{
			"errMsg": "Failed in BindJSON!",
		})
		return
	}
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
func setCouponsToRedis(username string, coupons Coupon) bool {
	return true
}

// 任务2
func getCouponsFromRedisOrDatabase(username string, coupons string) Coupon {
	return Coupon{}
}

// 任务3
func patchCoupons(c *gin.Context) {
	var err error
	tokenString := c.Request.Header.Get("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})
	// 5xx: 服务端错误
	if err != nil {
		c.JSON(504, gin.H{"errMsg": "Gateway Timeout"})
		return
	}
	//认证失败
	if err != nil || validateJWT(c) == false {
		c.JSON(401, gin.H{"errMsg": "Authorization Failed"})
		return
	}

	// userName: 用户名
	// sellerName: 商家名
	// couponName: 优惠券名
	// 从token.Claims获取用户名
	userName := token.Claims.(*MyClaims).Uname
	sellerName := c.Param("username")
	couponName := c.Param("name")
	// 204: 已经有了优惠券
	_, exists := hashset[userName]
	if exists {
		c.JSON(204, gin.H{"errMsg": "Already had the same coupon"})
		return
	}

	coupon := getCouponsFromRedisOrDatabase(sellerName, couponName)
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
	write := setCouponsToRedis(userName, coupon)
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
	err = mqueue.SendMessage(sellerName, couponName, uid, t.Unix())
	if err != nil {
		c.JSON(504, gin.H{"errMsg": "Gateway Timeout"})
		return
	}
	err, res := mqueue.ReceiveMessage(sellerName, couponName, uid, t.Unix())
	if err != nil {
		c.JSON(504, gin.H{"errMsg": "Gateway Timeout"})
		return
	}

	//返回0代表优惠券数目为0，返回2代表抢券成功，返回1代表用户已经抢到该券不可重复抢，返回-1代表数据库访问错误，返回-2代表超时
	switch res {
	case -2:
		c.JSON(504, gin.H{"errMsg": "Time out"})
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
