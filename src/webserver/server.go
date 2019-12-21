package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
)

// redis 默认是没有密码和使用0号db
var redis_client *redis.Client
var mysql_client *sql.DB

func init() {
	redis_client = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
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
	mysql_client, err = sql.Open("mysql", "root:Lichen1996@tcp(127.0.0.1:3306)/projectdb")
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
	Username    string  `json:"username"`
	Coupons     string  `json:"name"`
	Amount      int32   `json:"amount"`
	Stock       float32 `json:"stock"`
	Left        int32   `json:"left"`
	Description string  `json:"description"`
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

//---------------------------------------------------------------
//---------------------------------------------------------------
//任务2----------------------------------------------------------

/*
type CouponV2 struct {
	Username string `json:"username"`

	Coupons string `json:"name"`

	Amount int32 `json:"amount"`

	Stock float32 `json:"stock"`

	Left int32 `json:"left"`

	Description string `json:"description"`
}

func (c *CouponV2) ToCoupon() Coupon {
	var cou Coupon
	cou.username = c.Username
	cou.coupons = c.Coupons
	cou.amount = c.Amount
	cou.stock = c.Stock
	cou.left = c.Left
	cou.description = c.Description
	return cou
}
*/

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
	err := json.Unmarshal(resultJSON, &result)
	if err != nil {
		fmt.Println(err.Error())
	}
	c.JSON(200, result)
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

/*
func main() {
	// gin.SetMode(gin.ReleaseMode)
	router := setupRouter()
	err := router.Run(":8080")
	if err != nil {
		fmt.Println("Error starting server")
	}
}
*/
