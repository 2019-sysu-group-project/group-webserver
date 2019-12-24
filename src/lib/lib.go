package lib

import (
	"crypto/md5"
	"fmt"
	"io"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	Key []byte = []byte("JWT key of GXNXFD")
)

// MyClaims: Customer Claims
type MyClaims struct {
	Uname string `json:"username"`
	jwt.StandardClaims
}

// md5
func Md5Hash(data string) string {
	hash := md5.New()
	io.WriteString(hash, data)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// generate JWT token
func GenToken(username string) string {
	claims := &MyClaims{
		username,
		jwt.StandardClaims{
			NotBefore: int64(time.Now().Unix()),
			ExpiresAt: int64(time.Now().Unix() + 3600),
			Issuer:    "GXNXFD",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(Key)
	if err != nil {
		return ""
	}
	return ss
}
