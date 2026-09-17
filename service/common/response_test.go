package common

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestResponseError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		err  error
		want map[string]interface{}
	}{
		{
			name: "coded",
			err:  Coded("PORT_OCCUPIED", errors.New("port 1080 is already in use"), map[string]interface{}{"port": 1080}),
			want: map[string]interface{}{
				"code":      "FAIL",
				"message":   "port 1080 is already in use",
				"data":      nil,
				"errorCode": "PORT_OCCUPIED",
				"params":    map[string]interface{}{"port": float64(1080)},
			},
		},
		{
			name: "uncoded",
			err:  errors.New("plain error"),
			want: map[string]interface{}{
				"code":    "FAIL",
				"message": "plain error",
				"data":    nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			ResponseError(ctx, tt.err)

			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
			}
			var got map[string]interface{}
			if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("body = %#v, want %#v", got, tt.want)
			}
		})
	}
}
