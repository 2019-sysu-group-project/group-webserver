package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"

	"crypto/md5"
	"io"
	"strconv"
	"webserver/mqueue"

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
// var mysql_client =sql.Open("mysql", "root:123@tcp(127.0.0.1:13306)/projectdb")

// Coupon has username as shopper's name, coupons as its name
type Coupon struct {
	Username    string  `json:"username"`
	Coupons     string  `json:"name"`
	Amount      int32   `json:"amount"`
	Stock       float32 `json:"stock"`
	Left        int32   `json:"left"`
	Description string  `json:"description"`
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
	return err == nil
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
		if !isUserExist(json.Username) {
			c.JSON(401, gin.H{
				"kind":   "",
				"errMsg": "用户不存在",
			})
			return
		}
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
	return err == nil
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

//---------------------------------------------------------------
//---------------------------------------------------------------
//任务2----------------------------------------------------------

//将Coupon拼接成字符串，以#分隔：...#...#...
func (c *Coupon) ToString() string {
	var s string
	s = fmt.Sprintf("%s#%s#%d#%f#%d#%s", c.Username, c.Coupons, c.Amount,
		c.Stock, c.Left, c.Description)
	return s
}

//将字符串转换成Coupon
func (c *Coupon) ToCoupon(s string) {
	j := strings.LastIndex(s, "#")
	c.Description = s[j+1:]
	s = s[:j]
	j = strings.LastIndex(s, "#")
	Left_coupons, _ := strconv.ParseInt(s[j+1:], 10, 32)
	c.Left = int32(Left_coupons)
	s = s[:j]
	j = strings.LastIndex(s, "#")
	Stock, _ := strconv.ParseFloat(s[j+1:], 32)
	c.Stock = float32(Stock)
	s = s[:j]
	j = strings.LastIndex(s, "#")
	Amount, _ := strconv.ParseInt(s[j+1:], 10, 32)
	c.Amount = int32(Amount)
	s = s[:j]
	j = strings.LastIndex(s, "#")
	c.Coupons = s[j+1:]
	s = s[:j]
	c.Username = s
}

// 任务2
//check user if is business or not, exist or not
//0:buniess, 1:customer, 2:not exit
func checkUser(username string) int {
	var user User_DB
	err := mysql_client.QueryRow("SELECT username, password, kind FROM User WHERE username=?",
		username).Scan(&user.Username, &user.Password, &user.Kind)
	if err != nil { //未查询到的情况
		return 2
	} else if user.Kind == 0 {
		return 1
	} else {
		return 0
	}
}

// 任务2
func createCoupons(c *gin.Context) {
	if !validateJWT(c) {
		c.JSON(401, gin.H{"errMsg": "认证失败"})
	}
	var couponJSON Coupon
	err := c.BindJSON(&couponJSON)
	couponJSON.Username = c.Param("username")
	couponJSON.Left = couponJSON.Amount
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
	} else {
		flag := checkUser(couponJSON.Username)
		if flag == 2 {
			c.JSON(400, gin.H{"errMsg": "不存在的商家"})
		} else if flag == 1 {
			c.JSON(400, gin.H{"errMsg": "非商家不能创建优惠券"})
		} else {
			mysql_client.Query("INSERT INTO Coupon (username, coupons, amount, stock, left_coupons, description) VALUES (?,?,?,?,?,?)",
				couponJSON.Username, couponJSON.Coupons, couponJSON.Amount,
				couponJSON.Stock, couponJSON.Left, couponJSON.Description)
			c.JSON(201, gin.H{"errMsg": ""})
		}
	}
}

type GetCous struct {
	ErrMsg string   `json:"errMsg"`
	Data   []Coupon `json:"data"`
}

// 任务2
func getCouponsInformation(c *gin.Context) {
	Username := c.Param("username")
	page := c.Query("page")
	var resu GetCous
	var cou Coupon
	deviation, _ := strconv.Atoi(page)
	flag := checkUser(Username)
	if flag != 2 {
		query, _ := mysql_client.Query("SELECT username, coupons, amount, stock, left_coupons, description FROM Coupon WHERE username=? limit ?,20",
			Username, (deviation-1)*20)
		//	defer query.Close()
		for query.Next() {
			query.Scan(&cou.Username, &cou.Coupons, &cou.Amount,
				&cou.Stock, &cou.Left, &cou.Description)
			resu.Data = append(resu.Data, cou)
		}
		resultJSON, _ := json.Marshal(resu)
		var result map[string]interface{}
		json.Unmarshal(resultJSON, &result)
		c.JSON(200, result)
	} else {
		c.JSON(401, gin.H{"errMsg": "用户不存在"})
	}
}

// 任务2
func getCouponsFromRedis(Username string, cou string) (Coupon, error) {
	re, err := redis_client.Get(Username + "#" + cou).Result()
	var result Coupon
	if err == nil {
		result.ToCoupon(re)
	}
	return result, err
}

// 任务2
func setCouponsToRedis(Username string, cou Coupon) {
	redis_client.Set(Username+"#"+cou.Coupons, cou.ToString(), 2*time.Second)
}

// 任务2
func getCouponsFromRedisOrDatabase(Username string, cou string) (Coupon, error) {
	var result Coupon
	result, err := getCouponsFromRedis(Username, cou)
	if err != nil {
		query, err := mysql_client.Query("SELECT * FROM Coupon WHERE Username=? AND Coupons=?", Username, cou)
		if err == nil {
			defer query.Close()
			query.Next()
			query.Scan(&result.Username, &result.Coupons, &result.Amount,
				&result.Stock, &result.Left, &result.Description)
			setCouponsToRedis(Username, result)
		}
	}
	return result, err
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
	userName := token.Claims.(MyClaims).Uname
	sellerName := c.Param("username")
	couponName := c.Param("name")
	// 204: 已经有了优惠券
	_, exists := hashset[userName]
	if exists {
		c.JSON(204, gin.H{"errMsg": "Already had the same coupon"})
		return
	}

	coupon, err := getCouponsFromRedisOrDatabase(sellerName, couponName)
	// 5xx: 服务端错误
	if err != nil {
		c.JSON(504, gin.H{"errMsg": "Gateway Timeout"})
		return
	}
	// 204: 优惠券无库存
	if coupon.Left == 0 {
		c.JSON(204, gin.H{"errMsg": "The coupon is out of stock"})
		return
	}

	coupon.Left--
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
