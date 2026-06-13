package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/txbao/goeasy/contextx"
)

func TestInjectOperatorContextWithJWTClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var got contextx.OperatorContext
	engine := gin.New()
	engine.Use(requestID())
	engine.Use(func(c *gin.Context) {
		c.Set("jwt_subject", "admin-1")
		c.Set("jwt_customer_id", int64(100))
		c.Set("jwt_platform_admin", true)
		c.Next()
	})
	engine.Use(InjectOperatorContext())
	engine.GET("/ok", func(c *gin.Context) {
		got = contextx.OperatorFrom(c.Request.Context())
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	engine.ServeHTTP(w, req)

	if got.UserID != "admin-1" || got.LoginID != "admin-1" {
		t.Fatalf("user: %+v", got)
	}
	if got.CustomerID != 100 || !got.PlatformAdmin {
		t.Fatalf("tenant: %+v", got)
	}
	if got.RequestID == "" {
		t.Fatal("expected request_id")
	}
	if got.IP == "" {
		t.Fatal("expected client ip")
	}
	if got.DeviceInfo != "Mozilla/5.0" {
		t.Fatalf("device: %q", got.DeviceInfo)
	}
}

func TestInjectOperatorContextWithoutJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var got contextx.OperatorContext
	engine := gin.New()
	engine.Use(requestID())
	engine.Use(InjectOperatorContext())
	engine.GET("/ok", func(c *gin.Context) {
		got = contextx.OperatorFrom(c.Request.Context())
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	engine.ServeHTTP(w, req)

	if got.UserID != "" || got.CustomerID != 0 {
		t.Fatalf("expected empty jwt fields, got %+v", got)
	}
	if got.RequestID == "" || got.IP == "" {
		t.Fatalf("expected request/ip, got %+v", got)
	}
}

func TestTruncateDeviceInfo(t *testing.T) {
	long := strings.Repeat("x", 600)
	got := truncateDeviceInfo(long)
	if len(got) != maxDeviceInfoLen {
		t.Fatalf("len=%d", len(got))
	}
}
