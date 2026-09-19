package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/db"
	"github.com/v2rayA/v2rayA/pkg/server/jwt"
)

func TestLoginTokenSuppliesIdentity(t *testing.T) {
	token, err := Register("claim-test", "test-password")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Delete("accounts", "claim-test")
	r := gin.New()
	r.Use(jwt.JWTAuth(false))
	r.GET("/", func(c *gin.Context) { c.String(http.StatusOK, c.GetString("Name")) })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, req)
	if response.Code != http.StatusOK || response.Body.String() != "claim-test" {
		t.Fatalf("identity response: %d %q", response.Code, response.Body.String())
	}
}
