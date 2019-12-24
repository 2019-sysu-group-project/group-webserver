package model

import (
	"fmt"
	"strconv"
	"strings"
)

// User_DB not known yet
type User_DB struct {
	Username string
	Password string
	Kind     int
}

type GetCous struct {
	ErrMsg string   `json:"errMsg"`
	Data   []Coupon `json:"data"`
}

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

type RequestMessage struct {
	Username    string
	Coupon      string
	Uuid        string // 表示用户发起请求的唯一id
	RequestTime int64  // 用户发起请求的时间
	Result      int
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

//将Coupon拼接成字符串，以#分隔：...#...#...
func (c *Coupon) ToString() string {
	var s string
	s = fmt.Sprintf("%s#%s#%d#%f#%d#%s", c.Username, c.Coupons, c.Amount,
		c.Stock, c.Left, c.Description)
	return s
}
