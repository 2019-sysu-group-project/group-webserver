package controller

// hashset 存储元组(用户名, 商家名_优惠券名)
var hashset map[string]string

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Kind     string `json:"kind"`
}
