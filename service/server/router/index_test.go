package router

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterRoutesIncludesWorkflowEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	RegisterRoutes(engine)

	for _, route := range engine.Routes() {
		if route.Method == http.MethodPost && route.Path == "/api/workflow/refresh-subscription-and-reselect" {
			return
		}
	}
	t.Fatal("workflow route not registered")
}
