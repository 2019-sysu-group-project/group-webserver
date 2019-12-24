package controller

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"webserver.example/lib"
)

func ValidateJWT(c *gin.Context) bool {
	// 需要编写JWT的验证机制，作为其他人能调用的一部分
	token := c.Request.Header.Get("Authorization")
	_, err := jwt.Parse(token, func(*jwt.Token) (interface{}, error) {
		return lib.Key, nil
	})
	return err == nil
}
