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
	handler := NewReverseProxy(lbInstance, 5)

	req := httptest.NewRequest("GET", "http://proxy.example.com/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("期望状态码 200，得到 %d", rec.Code)
	}
	if rec.Body.String() != "backend response" {
		t.Errorf("期望响应体 'backend response'，得到 %s", rec.Body.String())
	}
}

// 测试所有后端不可用时返回 503
func TestReverseProxyAllUnhealthy(t *testing.T) {
	lbInstance := lb.New([]*url.URL{}) // 空列表，HealthyCount() == 0
	handler := NewReverseProxy(lbInstance, 5)

	req := httptest.NewRequest("GET", "http://proxy.example.com/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("期望状态码 503，得到 %d", rec.Code)
	}
}

// 测试有健康后端但转发失败（如后端超时）返回 502
func TestReverseProxyBackendTimeout(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // 延迟2秒
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	backendURL, _ := url.Parse(backend.URL)
	lbInstance := lb.New([]*url.URL{backendURL})
	handler := NewReverseProxy(lbInstance, 1) // 超时1秒

	req := httptest.NewRequest("GET", "http://proxy.example.com/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("期望超时后返回 502，得到 %d", rec.Code)
	}
}

// 测试有健康后端但后端返回 500 等错误（应透传给客户端，不应该是 502）
func TestReverseProxyBackendError(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("backend error"))
	}))
	defer backend.Close()

	backendURL, _ := url.Parse(backend.URL)
	lbInstance := lb.New([]*url.URL{backendURL})
	handler := NewReverseProxy(lbInstance, 5)

	req := httptest.NewRequest("GET", "http://proxy.example.com/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// 后端返回 500，代理应原样返回 500，而不是 502
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 500，得到 %d", rec.Code)
	}
	if rec.Body.String() != "backend error" {
		t.Errorf("期望响应体 'backend error'，得到 %s", rec.Body.String())
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
	handler := NewReverseProxy(lbInstance, 5)

	req := httptest.NewRequest("GET", "http://proxy.example.com/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("期望 200，得到 %d", rec.Code)
	}
}
