package model

import (
	"database/sql"
	"fmt"
	"time"

	"webserver.example/lib"
)

// check if the user already exists in DB
func IsUserExist(usernameQuery string) bool {
	var user User_DB
	err := Mysql_client.QueryRow("SELECT username, password, kind FROM User WHERE username=?", usernameQuery).Scan(&user.Username, &user.Password, &user.Kind)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

// insert user into DB
func InsertUser(username string, password string, kind int) bool {
	result, err := Mysql_client.Exec("INSERT INTO User(username, password, kind) VALUES(?,?,?)", username, password, kind)
	if err != nil {
		// insert failed
		return false
	}
	_, err = result.LastInsertId()
	if err != nil {
		return false
	}
	_, err = result.RowsAffected()
	return err == nil
}

// authenticate user from DB
func AuthenticateUser(username string, password string) bool {
	passwordHash := lib.Md5Hash(password)
	// passwordHash := password
	var user User
	err := Mysql_client.QueryRow("SELECT username, password FROM User WHERE username=?", username).Scan(&user.Username, &user.Password)
	if err == sql.ErrNoRows {
		return false
	} else {
		if user.Username == username && user.Password == passwordHash {
			return true
		}
	}
	return false
}

func CheckUser(username string) int {
	var user User_DB
	err := Mysql_client.QueryRow("SELECT username, password, kind FROM User WHERE username=?",
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
func SetCouponsToRedis(Username string, cou Coupon) {
	Redis_client.Set(Username+"#"+cou.Coupons, cou.ToString(), 2*time.Second)
}

// 任务2
func GetCouponsFromRedisOrDatabase(Username string, cou string) (Coupon, error) {
	var result Coupon
	result, err := GetCouponsFromRedis(Username, cou)
	if err != nil {
		query, err := Mysql_client.Query("SELECT * FROM Coupon WHERE Username=? AND Coupons=?", Username, cou)
		if err == nil {
			defer query.Close()
			query.Next()
			var id int
			query.Scan(&id, &result.Username, &result.Coupons, &result.Amount,
				&result.Stock, &result.Left, &result.Description)
			SetCouponsToRedis(Username, result)
			return result, nil
		}
	}
	return result, err
}

func CheckCouponNameAva(couponName string) bool {
	var coupon Coupon
	err := Mysql_client.QueryRow("SELECT coupons FROM Coupon WHERE coupons=?",
		couponName).Scan(&coupon.Coupons)
	if err == nil {
		return false
	}
	fmt.Println(err)
	return true
}

// 任务2
func GetCouponsFromRedis(Username string, cou string) (Coupon, error) {
	re, err := Redis_client.Get(Username + "#" + cou).Result()
	var result Coupon
	if err == nil {
		result.ToCoupon(re)
	}
	return result, err
}

// OccupyCoupon 检查redis字段，若用户数已满，则直接返回；否则先写入redis，再交给消息队列处理
func OccupyCoupon(coupon, username string) (bool, error) {
	return false, nil
}
