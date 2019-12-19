package main

import (
	// "net/http"
	"net/http/httptest"
	"testing"
	"encoding/json"
)

// 任务1
func testRegisterUser(t *testing.T){
	
}

// 任务1
func testUserLogin(t *testing.T){

}
//任务2
func testCreateCoupons(t *testing.T){
	router:=setupRouter()
	var cou CouponV2
	cou=CouponV2{"root","food",10,3.2,10,"gff"}
	strCou,_ := json.Marshal(cou)
	byCou:=bytes.NewReader(strCou)
	req:=httptest.NewRequest("POST","/api/users/:username/coupons",byCou)
	w:=httptest.NewRecorder()
	router.ServeHTTP(w,req)
	assert.Equal(t,http.StatusOK,w.Code)
}
//任务2
func testGetCouponsInformation(t *testing.T){
	router:=setupRouter()
	req:=httptest.NewRequest("GET","/api/users/:username/coupons?username=root&coupons=food",nil)
	w:=httptest.NewRecorder()
	router.ServeHTTP(w,req)
	assert.Equal(t,http.StatusOK,w.Code)
	var couv2 CouponV2
	json.Unmarshal([]byte(w.Body.String()), &couv2)
	var cou Coupon
	cou=couv2.ToCoupon()
	assert.Equal(t,"root",cou.username)
	assert.Equal(t,"food",cou.coupons)
}

// 任务3
func testPatchCoupons(t *testing.T){
	
}
