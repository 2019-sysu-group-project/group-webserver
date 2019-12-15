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
func testPatchCoupons(t *testing.T) {
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("Patch", "/api/users/:username/coupons/:name", nil)
	
}
