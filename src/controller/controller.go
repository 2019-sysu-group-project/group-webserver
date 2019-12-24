package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"webserver.example/lib"
	"webserver.example/model"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func RegisterUser(c *gin.Context) {
	var json model.User

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
		if model.IsUserExist(username) { // user already exists
			c.JSON(400, gin.H{
				"errMsg": "用户已存在",
			})
			return
		}
		passwordHash := lib.Md5Hash(password)
		// insert user to DB
		if model.InsertUser(username, passwordHash, kindInt) {
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

func UserLogin(c *gin.Context) {
	var json model.User
	if c.BindJSON(&json) == nil {
		if json.Username == "" || json.Password == "" {
			c.JSON(401, gin.H{
				"kind":   "",
				"errMsg": "空用户或空密码",
			})
			return
		}
		if !model.IsUserExist(json.Username) {
			c.JSON(401, gin.H{
				"kind":   "",
				"errMsg": "用户不存在",
			})
			return
		}
		if model.AuthenticateUser(json.Username, json.Password) {
			token := lib.GenToken(json.Username)
			if token == "" {
				c.JSON(401, gin.H{
					"kind":   "",
					"errMsg": "Generate token failed.",
				})
				return
			}
			var user model.User_DB
			err := model.Mysql_client.QueryRow("SELECT kind FROM User WHERE username=?", json.Username).Scan(&user.Kind)
			if err != nil {
				c.JSON(500, gin.H{
					"kind":   "",
					"errMsg": "Query DB failed.",
				})
				return
			}
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
				"Authorization": token,
				"kind":          kindString,
				"errMsg":        "",
			})
			return
		}
		c.JSON(401, gin.H{
			"kind":   "",
			"errMsg": "错误密码",
		})
		return
	}
	c.JSON(400, gin.H{
		"errMsg": "获取json数据失败",
	})
	return
}

func CreateCoupons(c *gin.Context) {
	if !ValidateJWT(c) {
		c.JSON(401, gin.H{"errMsg": "认证失败"})
	}
	var couponJSON model.Coupon
	err := c.BindJSON(&couponJSON)
	couponJSON.Username = c.Param("username")
	couponJSON.Left = couponJSON.Amount
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
	} else {
		flag := model.CheckUser(couponJSON.Username)
		if flag == 2 {
			c.JSON(400, gin.H{"errMsg": "不存在的商家"})
		} else if flag == 1 {
			c.JSON(400, gin.H{"errMsg": "非商家不能创建优惠券"})
		} else {
			availability := model.CheckCouponNameAva(couponJSON.Coupons)
			if availability == false {
				c.JSON(400, gin.H{"errMsg": "相同名字的优惠券已存在"})
				return
			}
			model.Mysql_client.Query("INSERT INTO Coupon (username, coupons, amount, stock, left_coupons, description) VALUES (?,?,?,?,?,?)",
				couponJSON.Username, couponJSON.Coupons, couponJSON.Amount,
				couponJSON.Stock, couponJSON.Left, couponJSON.Description)
			c.JSON(201, gin.H{"errMsg": ""})
		}
	}
}

func PatchCoupons(c *gin.Context) {
	var err error
	tokenString := c.Request.Header.Get("Authorization")
	token, err := jwt.ParseWithClaims(tokenString, &lib.MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return lib.Key, nil
	})
	// 5xx: 服务端错误
	if err != nil {
		log.Println(err)
		c.JSON(504, gin.H{"errMsg": "server error"})
		return
	}
	//认证失败
	if ValidateJWT(c) == false {
		c.JSON(401, gin.H{"errMsg": "Authorization Failed"})
		return
	}

	// userName: 用户名
	// sellerName: 商家名
	// couponName: 优惠券名
	// 从token.Claims获取用户名
	userName := token.Claims.(*lib.MyClaims).Uname
	sellerName := c.Param("username")
	couponName := c.Param("name")
	// 204: 已经有了优惠券
	_, exists := hashset[userName]
	if exists {
		c.JSON(204, gin.H{"errMsg": "Already had the same coupon"})
		return
	}
	// redis这部分可能有bug!!!(start)
	coupon, err := model.GetCouponsFromRedisOrDatabase(sellerName, couponName)
	// 5xx: 服务端错误
	if err != nil {
		log.Println(err)
		c.JSON(504, gin.H{"errMsg": "Gateway Timeout"})
		return
	}
	// 204: 优惠券无库存
	if coupon.Left == 0 {
		c.JSON(204, gin.H{"errMsg": "The coupon is out of stock"})
		return
	}

	// TODO: 1. 使用pipeline优化redis使用
	//       2. 检查已被抢到的数量，超过则直接返回抢光
	//       3. 没抢光检查是否在set中，是直接返回
	//       4.
	coupon.Left--
	model.SetCouponsToRedis(userName, coupon)
	// 5xx: 服务端错误
	if err != nil {
		c.JSON(504, gin.H{"errMsg": "Gateway Timeout"})
		return
	}
	// redis这部分可能有bug!!!(end)

	// 将用户请求转发到消息队列中，等待消息队列对mysql进行操作并返回结果
	t := time.Now()
	// 生成uuid
	u := uuid.NewV4()
	uid := u.String()
	// 先判断是否能成功发送消息
	err = model.SendMessage(userName, couponName, uid, t.Unix())
	if err != nil {
		c.JSON(504, gin.H{"errMsg": "Gateway Timeout"})
		return
	}
	err, res := model.ReceiveMessage(userName, couponName, uid, t.Unix())
	if err != nil {
		c.JSON(504, gin.H{"errMsg": "Gateway Timeout"})
		return
	}
	fmt.Println(res)

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

func GetCouponsInformation(c *gin.Context) {
	Username := c.Param("username")
	page := c.Query("page")
	if !ValidateJWT(c) && model.CheckUser(Username) == 1 {
		c.JSON(401, gin.H{
			"errMsg": "认证错误",
		})
	}
	var resu model.GetCous
	var cou model.Coupon
	deviation, _ := strconv.Atoi(page)
	flag := model.CheckUser(Username)
	if flag != 2 {
		query, _ := model.Mysql_client.Query("SELECT username, coupons, amount, stock, left_coupons, description FROM Coupon WHERE username=? limit ?,20",
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
