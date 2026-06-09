package apisign

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/txbao/goeasy/config"
)

// Verifier RSA2（SHA256withRSA）验签器。
type Verifier struct {
	skew   time.Duration
	apps   map[string]*rsa.PublicKey
}

func New(cfg config.APISignCfg) (*Verifier, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	v := &Verifier{
		skew: time.Duration(cfg.TimestampSkewMs) * time.Millisecond,
		apps: make(map[string]*rsa.PublicKey, len(cfg.Apps)),
	}
	for appID, app := range cfg.Apps {
		pub, err := parsePublicKey(app.PublicKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("api_sign app %s: %w", appID, err)
		}
		v.apps[appID] = pub
	}
	return v, nil
}

func parsePublicKey(pemOrB64 string) (*rsa.PublicKey, error) {
	raw := strings.TrimSpace(pemOrB64)
	if raw == "" {
		return nil, errors.New("empty public key")
	}
	if strings.Contains(raw, "BEGIN") {
		block, _ := pem.Decode([]byte(raw))
		if block == nil {
			return nil, errors.New("invalid PEM")
		}
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rsaPub, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("not RSA public key")
		}
		return rsaPub, nil
	}
	der, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}
	pub, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, err
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not RSA public key")
	}
	return rsaPub, nil
}

// AuthHeader 解析 Authorization 头字段。
type AuthHeader struct {
	Algorithm string
	AppID     string
	Nonce     string
	Timestamp string
	Signature string
}

// ParseAuthorization 解析 `algorithm=RSA2,appid=...,nonce_str=...,timestamp=...,signature=...`。
func ParseAuthorization(raw string) (AuthHeader, error) {
	var h AuthHeader
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return h, errors.New("empty authorization")
	}
	parts := strings.Split(raw, ",")
	for _, p := range parts {
		kv := strings.SplitN(strings.TrimSpace(p), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch strings.ToLower(kv[0]) {
		case "algorithm":
			h.Algorithm = kv[1]
		case "appid":
			h.AppID = kv[1]
		case "nonce_str":
			h.Nonce = kv[1]
		case "timestamp":
			h.Timestamp = kv[1]
		case "signature":
			h.Signature = kv[1]
		}
	}
	if h.Algorithm == "" || h.AppID == "" || h.Nonce == "" || h.Timestamp == "" || h.Signature == "" {
		return h, errors.New("incomplete authorization")
	}
	return h, nil
}

// CanonicalQuery 规范化 Query（与 api-sign.md 一致）。
func CanonicalQuery(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}
	vals, err := url.ParseQuery(rawQuery)
	if err != nil {
		return rawQuery
	}
	keys := make([]string, 0, len(vals))
	for k := range vals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var pairs []string
	for _, k := range keys {
		vs := vals[k]
		sort.Strings(vs)
		for _, v := range vs {
			pairs = append(pairs, url.QueryEscape(k)+"="+url.QueryEscape(v))
		}
	}
	return strings.Join(pairs, "&")
}

// SigningPath 签名用 URL 路径（含规范化 Query）。
func SigningPath(r *http.Request) string {
	path := r.URL.Path
	cq := CanonicalQuery(r.URL.RawQuery)
	if cq != "" {
		return path + "?" + cq
	}
	return path
}

// BuildSignString 构造待签名字符串。
func BuildSignString(method, signPath, timestamp, nonce, body string) string {
	return strings.ToUpper(method) + "\n" +
		signPath + "\n" +
		timestamp + "\n" +
		nonce + "\n" +
		body + "\n"
}

// VerifyRequest 校验 HTTP 请求签名。
func (v *Verifier) VerifyRequest(r *http.Request, body []byte) error {
	if v == nil {
		return errors.New("api_sign disabled")
	}
	h, err := ParseAuthorization(r.Header.Get("Authorization"))
	if err != nil {
		return err
	}
	if !strings.EqualFold(h.Algorithm, "RSA2") {
		return errors.New("unsupported algorithm")
	}
	pub, ok := v.apps[h.AppID]
	if !ok {
		return errors.New("unknown appid")
	}
	ts, err := strconv.ParseInt(h.Timestamp, 10, 64)
	if err != nil {
		return errors.New("invalid timestamp")
	}
	now := time.Now().UnixMilli()
	if v.skew > 0 && (now-ts > int64(v.skew.Milliseconds()) || ts-now > int64(v.skew.Milliseconds())) {
		return errors.New("timestamp skew")
	}
	signStr := BuildSignString(r.Method, SigningPath(r), h.Timestamp, h.Nonce, string(body))
	sig, err := base64.StdEncoding.DecodeString(h.Signature)
	if err != nil {
		return err
	}
	hash := sha256.Sum256([]byte(signStr))
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, hash[:], sig); err != nil {
		return errors.New("invalid signature")
	}
	return nil
}
