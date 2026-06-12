package response

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	zerr "github.com/txbao/goeasy/errors"
	"github.com/txbao/goeasy/logger"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func resetOpts() {
	Configure(Options{LogServerErrors: true})
}

func testLogger(buf *bytes.Buffer) *logger.Logger {
	h := slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	return logger.NewFromSlog(slog.New(h))
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder) body {
	t.Helper()
	var b body
	if err := json.Unmarshal(w.Body.Bytes(), &b); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	return b
}

func TestFailBiz_JSON(t *testing.T) {
	resetOpts()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/items", nil)

	FailBiz(c, http.StatusBadRequest, 100001, "bad param")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("http status = %d", w.Code)
	}
	b := decodeBody(t, w)
	if b.Code != 100001 || b.Msg != "bad param" {
		t.Fatalf("body = %+v", b)
	}
}

func TestFailInternal_ProdMasksDetail(t *testing.T) {
	resetOpts()
	Configure(Options{Env: "prod", LogServerErrors: false})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/x", nil)

	FailInternal(c, errors.New("secret sql error"))

	b := decodeBody(t, w)
	if b.Code != int(zerr.BizCodeInternal) {
		t.Fatalf("code = %d", b.Code)
	}
	if b.Msg != "服务内部错误" {
		t.Fatalf("msg = %q, want masked", b.Msg)
	}
}

func TestFailInternal_DevExposesDetail(t *testing.T) {
	resetOpts()
	Configure(Options{Env: "dev", LogServerErrors: false})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/x", nil)

	FailInternal(c, errors.New("debug detail"))

	b := decodeBody(t, w)
	if b.Msg != "debug detail" {
		t.Fatalf("msg = %q", b.Msg)
	}
}

func TestFailInternal_LogsOn500(t *testing.T) {
	resetOpts()
	var buf bytes.Buffer
	Configure(Options{
		Env:             "prod",
		Logger:          testLogger(&buf),
		LogServerErrors: true,
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/fail", nil)
	c.Set("request_id", "req-1")

	FailInternal(c, errors.New("boom"))

	if !bytes.Contains(buf.Bytes(), []byte("http_server_error")) {
		t.Fatalf("expected error log, got: %s", buf.String())
	}
	if !bytes.Contains(buf.Bytes(), []byte("req-1")) {
		t.Fatalf("expected request_id in log, got: %s", buf.String())
	}
	if !bytes.Contains(buf.Bytes(), []byte("500001")) {
		t.Fatalf("expected biz_code in log, got: %s", buf.String())
	}
}

func TestFailErr_BizError(t *testing.T) {
	resetOpts()
	Configure(Options{LogServerErrors: false})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)

	FailErr(c, zerr.AuthMissing("no token"))

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", w.Code)
	}
	b := decodeBody(t, w)
	if b.Code != int(zerr.BizCodeAuthMissing) {
		t.Fatalf("code = %d", b.Code)
	}
}

func TestFailErr_CodedError(t *testing.T) {
	resetOpts()
	Configure(Options{LogServerErrors: false})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)

	FailErr(c, zerr.NotFound("missing"))

	b := decodeBody(t, w)
	if w.Code != http.StatusNotFound || b.Code != http.StatusNotFound {
		t.Fatalf("http=%d body=%+v (CodedError 仍走 deprecated Fail)", w.Code, b)
	}
}

func TestFailErr_FallbackInternal(t *testing.T) {
	resetOpts()
	Configure(Options{Env: "prod", LogServerErrors: false})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)

	FailErr(c, errors.New("raw"))

	b := decodeBody(t, w)
	if b.Code != int(zerr.BizCodeInternal) {
		t.Fatalf("code = %d", b.Code)
	}
}
