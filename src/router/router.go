package router

import (
	"github.com/gin-gonic/gin"
	"webserver.example/controller"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	// task1
	router.POST("/api/users", controller.RegisterUser)
	router.POST("/api/auth", controller.UserLogin)
	// task2
	router.POST("/api/users/:username/coupons", controller.CreateCoupons)
	router.GET("/api/users/:username/coupons", controller.GetCouponsInformation)
	// task3
	router.PATCH("/api/users/:username/coupons/:name", controller.PatchCoupons)

	// used for testing
	// router.GET("/validate", controller.testValidateJWT)
	// router.GET("/test", controller.testMyFunc)
	return router
}
