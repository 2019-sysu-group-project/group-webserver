package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

type Coupon struct {
	username string
	coupons string
	amount int32
	stock float32
	left int32
	description string
}

// 任务1
func registerUser(c *gin.Context){
	fmt.Println("This is registerUser")
}

// 任务1
func validateJWT() bool{
	// 需要编写JWT的验证机制，作为其他人能调用的一部分
	return true
}

// 任务1
func userLogin(c *gin.Context){
	
}

// 任务2
func createCoupons(c *gin.Context){
	
}

// 任务2
func getCouponsInformation(c *gin.Context){
	
}

// 任务2
func getCouponsFromRedisOrDatabase(username string, coupons string) Coupon{
	return Coupon{}
}

// 任务3
func setCouponsToRedisAndDatabase(coupon Coupon) bool{
	// true set成功，false set失败
	return true
}

// 任务3
func patchCoupons(c *gin.Context){
	
}

func main()  {
	// gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	
	router.PATCH("/api/users/:username/coupons/:name", patchCoupons)
	router.POST("/api/users", registerUser)
	
	router.POST("/api/auth", userLogin)
	router.POST("/api/users/:username/coupons", createCoupons)
	
	router.GET("/api/users/:username/coupons", getCouponsInformation)

	err := router.Run(":8080")
	if err != nil{
		fmt.Println("Error starting server")
	}
}