package main

import (
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"log"
	"testing"
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
func testPatchCoupons(t *testing.T) (int, int64) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("Patch", "/api/users/:username/coupons/:name", nil)
	router.ServeHTTP(w, req)
	// 结束
	defer func() {
		err := w.Result().Body.Close()
		if err != nil {
			log.Println(err)
		}
	}()
	var err5XX int
	var time int64 = 0
	if w.Result().StatusCode/500 > 0 {
		err5XX = 1
	} else {
		err5XX = 0
	}
	assert.Equal(t, 504, w.Code)
	return err5XX, time
}
