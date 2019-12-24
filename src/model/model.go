package model

import (
	"fmt"
	"time"

	"webserver.example/lib"
)

// check if the user already exists in DB
func IsUserExist(usernameQuery string) bool {
	var user User
	query := GormDB.Where("username = ?", usernameQuery).Find(&user)
	if query.Error != nil {
		fmt.Println(query.Error)
		return false
	}
	return true
}

// insert user into DB
func InsertUser(username string, password string, kind int) bool {
	user := User{Username: username, Password: password, Kind: kind}

	query := GormDB.Create(&user)
	if query.Error != nil {
		return false
	}
	return true
}

func CreateCoupon(coupon CouponInfo) error {
	query := GormDB.Create(&coupon)
	if query.Error != nil {
		return query.Error
	}
	return nil
}

// authenticate user from DB
func AuthenticateUser(username string, password string) bool {
	passwordHash := lib.Md5Hash(password)
	// passwordHash := password
	var user User
	query := GormDB.Where("username = ?", username).Find(&user)
	if query.Error != nil {
		return false
	}
	if user.Username == username && user.Password == passwordHash {
		return true
	}
	return false
}

func CheckUser(username string) int {
	var user User
	query := GormDB.Where("username = ?", username).Find(&user)
	if query.Error != nil {
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
		var coupon CouponInfo
		query := GormDB.Where("username = ? AND coupons = ?", Username, cou).Find(&coupon)
		if query.Error != nil {
			return result, query.Error
		}
		result.Username = coupon.Username
		result.Coupons = coupon.Coupons
		result.Amount = int32(coupon.Amount)
		result.Stock = float32(coupon.Stock)
		result.Left = int32(coupon.Left)
		result.Description = coupon.Description
		SetCouponsToRedis(Username, result)
		return result, nil
	}
	return result, err
}

func CheckCouponNameAva(couponName string) bool {
	var coupon CouponInfo
	query := GormDB.Where("coupons = ?", couponName).Find(&coupon)
	if query.RecordNotFound() {
		return true
	}
	if query.Error != nil {
		fmt.Println(query.Error)
		return false
	}
	return false
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

func GetCoupons(username string) ([]CouponInfo, error) {
	ret := make([]CouponInfo, 0)
	query := GormDB.Where("username = ?", username).Model(&CouponInfo{}).Find(&ret)
	if query.RecordNotFound() {
		return ret, nil
	}
	return ret, query.Error
}

// OccupyCoupon 检查redis字段，若用户数已满，则直接返回；否则先写入redis，再交给消息队列处理
func OccupyCoupon(coupon, username string) (bool, error) {
	return false, nil
}
