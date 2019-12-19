package main

import (
	// "net/http"
	// "net/http/httptest"
	"testing"
	"net/http/httptest"
	"net/http"
	"github.com/gin-gonic/gin"

)

// 任务1
func testRegisterUser(t *testing.T) {

}

// 任务1
func testUserLogin(t *testing.T) {

}

// 任务2
func testCreateCoupons(t *testing.T) {

}

// 任务2
func testGetCouponsInformation(t *testing.T) {

}

// 任务3
// @Return
// Param1: 返回1，若异常状态为5xx；否则返回0
// Param2: 返回响应所耗时间
// 待后续并发进程调用
func testPatchCoupons(t *testing.T) (int, int) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("Patch", "/api/users/:username/coupons/:name", nil)
	router.serveHTTP(w, req)
	 // 提取响应
	 result := w.Result()
	 defer result.Body.Close()
	 // 读取响应body
	 body,_ := ioutil.ReadAll(result.Body)
	var x int
	if code / 500 > 0 {
		x = 1
	} else {
		x = 0
	}
	body
}
