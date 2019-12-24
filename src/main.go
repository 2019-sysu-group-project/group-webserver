package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"webserver.example/router"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := router.SetupRouter()
	fmt.Println("Server started")
	err := router.Run(":8080")
	if err != nil {
		fmt.Println("Error starting server")
	}
}
