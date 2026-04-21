package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// 测试健康检查：所有后端都健康
func TestCheckerAllHealthy(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server2.Close()

	u1, _ := url.Parse(server1.URL)
	u2, _ := url.Parse(server2.URL)
	backendURLs := []*url.URL{u1, u2}

	var receivedHealthy [][]*url.URL
	onHealthy := func(healthy []*url.URL) {
		copy := make([]*url.URL, len(healthy))
		for i, u := range healthy {
			copy[i] = u
		}
		receivedHealthy = append(receivedHealthy, copy)
	}

	checker := NewChecker(backendURLs, "/health", 1, 1, onHealthy)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	checker.Start(ctx)

	time.Sleep(2500 * time.Millisecond)
	cancel()

	if len(receivedHealthy) == 0 {
		t.Fatal("回调从未被调用")
	}
	lastHealthy := receivedHealthy[len(receivedHealthy)-1]
	if len(lastHealthy) != 2 {
		t.Errorf("期望健康后端数量为 2，实际 %d", len(lastHealthy))
	}
}

// 测试健康检查：部分后端不健康
func TestCheckerPartialHealthy(t *testing.T) {
	healthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer healthyServer.Close()

	unhealthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer unhealthyServer.Close()

	uHealthy, _ := url.Parse(healthyServer.URL)
	uUnhealthy, _ := url.Parse(unhealthyServer.URL)
	backendURLs := []*url.URL{uHealthy, uUnhealthy}

	var healthyList []*url.URL
	onHealthy := func(healthy []*url.URL) {
		healthyList = healthy
	}

	checker := NewChecker(backendURLs, "/health", 1, 1, onHealthy)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	checker.Start(ctx)

	time.Sleep(1500 * time.Millisecond)
	cancel()

	if len(healthyList) != 1 {
		t.Errorf("期望 1 个健康后端，实际 %d", len(healthyList))
	}
	if healthyList[0].Host != uHealthy.Host {
		t.Errorf("健康后端应为 %s，实际 %s", uHealthy.Host, healthyList[0].Host)
	}
}

// 测试健康检查：所有后端都不可达
func TestCheckerAllUnreachable(t *testing.T) {
	u1, _ := url.Parse("http://127.0.0.1:19999")
	u2, _ := url.Parse("http://127.0.0.1:19998")
	backendURLs := []*url.URL{u1, u2}

	var healthyList []*url.URL
	onHealthy := func(healthy []*url.URL) {
		healthyList = healthy
	}

	checker := NewChecker(backendURLs, "/health", 1, 1, onHealthy)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	checker.Start(ctx)

	time.Sleep(1500 * time.Millisecond)
	cancel()

	if len(healthyList) != 0 {
		t.Errorf("期望 0 个健康后端，实际 %d", len(healthyList))
	}
}

// 测试回调为 nil 时不会 panic
func TestCheckerNilCallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	u, _ := url.Parse(server.URL)
	backendURLs := []*url.URL{u}

	checker := NewChecker(backendURLs, "/health", 1, 1, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	checker.Start(ctx)

	time.Sleep(1500 * time.Millisecond)
	cancel()
	// 没有 panic 即测试通过
}

// 基准测试：单次 checkAndUpdate 性能
func BenchmarkCheckAndUpdate(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	u, _ := url.Parse(server.URL)
	backendURLs := []*url.URL{u}
	checker := NewChecker(backendURLs, "/health", 1, 1, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.checkAndUpdate()
	}
}
