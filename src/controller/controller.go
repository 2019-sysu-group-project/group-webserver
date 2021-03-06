package controller

import (
	"log"
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
			kind := model.CheckUser(json.Username)
			if kind == 2 {
				c.JSON(500, gin.H{
					"kind":   "",
					"errMsg": "Query DB failed.",
				})
				return
			}
			var kindString string
			if kind == 0 {
				kindString = "customer"
			} else if kind == 1 {
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
		return
	}
	flag := model.CheckUser(couponJSON.Username)
	if flag == 2 {
		c.JSON(400, gin.H{"errMsg": "不存在的商家"})
		return
	}
	if flag == 1 {
		c.JSON(400, gin.H{"errMsg": "非商家不能创建优惠券"})
		return
	}
	availability := model.CheckCouponNameAva(couponJSON.Coupons)
	if availability == false {
		c.JSON(400, gin.H{"errMsg": "相同名字的优惠券已存在"})
		return
	}
	err = model.CreateCoupon(model.CouponInfo{
		Username:    couponJSON.Username,
		Coupons:     couponJSON.Coupons,
		Amount:      int(couponJSON.Amount),
		Stock:       int(couponJSON.Stock),
		Left:        int(couponJSON.Left),
		Description: couponJSON.Description,
	})
	if err != nil {
		c.JSON(500, gin.H{"errMsg": err})
	}
	// 设定商家的优惠券数目到redis中
	err = model.SetCouponsAmountOfMerchant(couponJSON.Username, couponJSON.Coupons, int(couponJSON.Amount))
	if err != nil {
		c.JSON(500, gin.H{"errMsg": err})
	}
	c.JSON(201, gin.H{"errMsg": ""})
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
	// sellerName := c.Param("username")
	couponName := c.Param("name")
	// 204: 已经有了优惠券
	exName, exists := hashset[userName]
	if exists && exName == couponName {
		c.JSON(204, gin.H{"errMsg": "Already had the same coupon"})
		return
	}

	// 从redis中扣库存
	cnt, err := model.OccupyCoupon(couponName, userName)
	if err != nil {
		c.JSON(500, gin.H{"errMsg": err})
		return
	}
	if cnt < 0 {
		c.JSON(204, gin.H{"errMsg": "优惠券已抢光"})
		return
	}

	// 将用户请求转发到消息队列中
	t := time.Now()
	// 生成uuid
	u := uuid.NewV4()
	uid := u.String()
	// 先判断是否能成功发送消息
	err = model.SendMessage(userName, couponName, uid, t.Unix())
	if err != nil {
		model.RollBackCoupon(couponName, userName)
		c.JSON(504, gin.H{"errMsg": "Gateway Timeout"})
		return
	}
	c.JSON(201, gin.H{})
}

// TODO: add page
func GetCouponsInformation(c *gin.Context) {
	Username := c.Param("username")
	// page := c.Query("page")
	if !ValidateJWT(c) && model.CheckUser(Username) == 1 {
		c.JSON(401, gin.H{
			"errMsg": "认证错误",
		})
	}
	// deviation, _ := strconv.Atoi(page)
	flag := model.CheckUser(Username)
	if flag != 2 {
		result, err := model.GetAllCouponsFromRedisByUsername(Username)
		//result, err := model.GetCoupons(Username)
		if err != nil {
			c.JSON(500, gin.H{"errMsg": err})
			return
		}
		c.JSON(200, result)
	} else {
		c.JSON(401, gin.H{"errMsg": "用户不存在"})
	}
}
