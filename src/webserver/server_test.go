package main

import (
	// "net/http"
	// "net/http/httptest"
	"testing"
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
	cou:=Coupon{"root","food",10,3.2,10,""}
	str_cou,_ := json.Marshal(cou)
	req:=httptest.NewRequest("POST","/api/users/:username/coupons",bytes.NewBuffer(str_cou))
	w:=httptest.NewRecorder()
	router.ServeHTTP(w,req)
	assert.Equal(t,http.StatusOK,w.Code)
}
//任务2
func testGetCouponsInformation(t*testing.T){
	router:=setupRouter()
	req:=httptest.NewRequest("GET","/api/users/:username/coupons",bytes.NewBuffer(str_cou))
	w:=httptest.NewRecorder()
	router.ServeHTTP(w,req)
	assert.Equal(t,http.StatusOK,w.Code)
	var cou Coupon
	err := json.Unmarshal([]byte(w.Body.String()), &cou)
	assert.Equal(t,"",cou.username)
	assert.Equal(t,"",cou.coupons)
}

// 任务3
func testPatchCoupons(t *testing.T){
	
}
