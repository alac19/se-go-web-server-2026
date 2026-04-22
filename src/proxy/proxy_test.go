package proxy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"se-go-web-server-2026/src/lb"
)

// 测试反向代理基本转发功能
func TestReverseProxyBasic(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend response"))
	}))
	defer backend.Close()

	backendURL, _ := url.Parse(backend.URL)
	lbInstance := lb.New([]*url.URL{backendURL})
	proxy := NewReverseProxy(lbInstance, 5)

	req := httptest.NewRequest("GET", "http://proxy.example.com/test", nil)
	rec := httptest.NewRecorder()

	proxy.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("期望状态码 200，得到 %d", rec.Code)
	}
	if rec.Body.String() != "backend response" {
		t.Errorf("期望响应体 'backend response'，得到 %s", rec.Body.String())
	}
}

// 测试没有可用后端时返回 502
func TestReverseProxyNoBackend(t *testing.T) {
	lbInstance := lb.New([]*url.URL{}) // 空列表
	proxy := NewReverseProxy(lbInstance, 5)

	req := httptest.NewRequest("GET", "http://proxy.example.com/test", nil)
	rec := httptest.NewRecorder()

	proxy.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("期望状态码 502，得到 %d", rec.Code)
	}
}

// 测试后端超时（ResponseHeaderTimeout）
func TestReverseProxyTimeout(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // 延迟2秒
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	backendURL, _ := url.Parse(backend.URL)
	lbInstance := lb.New([]*url.URL{backendURL})
	proxy := NewReverseProxy(lbInstance, 1) // 超时1秒

	req := httptest.NewRequest("GET", "http://proxy.example.com/test", nil)
	rec := httptest.NewRecorder()

	proxy.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("期望超时后返回 502，得到 %d", rec.Code)
	}
}

// 测试请求头传递（如 X-Forwarded-For）
func TestReverseProxyHeaders(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Forwarded-For") == "" {
			t.Error("X-Forwarded-For 头部未被设置")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	backendURL, _ := url.Parse(backend.URL)
	lbInstance := lb.New([]*url.URL{backendURL})
	proxy := NewReverseProxy(lbInstance, 5)

	req := httptest.NewRequest("GET", "http://proxy.example.com/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	rec := httptest.NewRecorder()

	proxy.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("期望 200，得到 %d", rec.Code)
	}
}
