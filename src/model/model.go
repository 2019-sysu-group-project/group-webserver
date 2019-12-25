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
func SetCouponsToRedis(Username string, cou Coupon) error {
	_, err := Redis_client.Set(Username+"#"+cou.Coupons, cou.ToString(), 2*time.Second).Result()
	if err != nil {
		return err
	}
	return nil
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
		err = SetCouponsToRedis(Username, result)
		if err != nil {
			return result, err
		}
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

// 每当mysql数据库添加新的优惠券的时候，需要将商家拥有的优惠券的列表更新到redis
func AddCouponsToList(username, couponName string) error {
	_, err := Redis_client.RPush(username, couponName).Result()
	if err != nil {
		return err
	}
	return nil
}

// 商家通过查询redis(若redis没有则查询mysql)获取自己发布过的所有的优惠券信息
func GetAllCouponsFromRedisByUsername(username string) ([]CouponInfo, error) {
	res := make([]CouponInfo, 0)
	len, err := Redis_client.LLen(username).Result() //返回名称为key的list的长度
	if err != nil {
		return res, err
	}
	re, err := Redis_client.LRange(username, 0, len).Result()
	if err != nil {
		return res, err
	}
	for _, v := range re {
		//resTmp, err := GetCouponsFromRedis(username, string(v))
		resTmp, err := GetCouponsFromRedisOrDatabase(username, v)
		if err != nil {
			return res, err
		}
		var resInfo CouponInfo
		resInfo.Username = resTmp.Username
		resInfo.Coupons = resTmp.Coupons
		resInfo.Amount = int(resTmp.Amount)
		resInfo.Stock = int(resTmp.Stock)
		resInfo.Left = int(resTmp.Left)
		resInfo.Description = resTmp.Description
		res = append(res, resInfo)
	}
	return res, nil

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
