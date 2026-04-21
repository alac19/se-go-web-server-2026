package lb

import (
	"net/url"
	"sync"
	"testing"
)

// 辅助函数：创建测试用的 URL 列表
func testBackends() []*url.URL {
	return []*url.URL{
		{Scheme: "http", Host: "127.0.0.1:9091"},
		{Scheme: "http", Host: "127.0.0.1:9092"},
		{Scheme: "http", Host: "127.0.0.1:9093"},
	}
}

// 测试空列表时 Next 返回 nil
func TestLoadBalancerEmpty(t *testing.T) {
	lb := New([]*url.URL{})
	if got := lb.Next(); got != nil {
		t.Errorf("空列表时 Next() = %v，期望 nil", got)
	}
}

// 测试单个后端：每次都应该返回同一个 URL
func TestLoadBalancerSingle(t *testing.T) {
	backends := []*url.URL{{Scheme: "http", Host: "single:8080"}}
	lb := New(backends)

	for i := 0; i < 5; i++ {
		got := lb.Next()
		if got == nil {
			t.Fatal("Next() 返回 nil，期望非 nil")
		}
		if got.Host != "single:8080" {
			t.Errorf("第 %d 次调用返回 Host = %s，期望 single:8080", i, got.Host)
		}
	}
}

// 测试轮询算法：按顺序循环返回
func TestLoadBalancerRoundRobin(t *testing.T) {
	backends := testBackends()
	lb := New(backends)

	expected := []string{
		"127.0.0.1:9091",
		"127.0.0.1:9092",
		"127.0.0.1:9093",
		"127.0.0.1:9091", // 第4次回到第一个
		"127.0.0.1:9092",
	}

	for i, want := range expected {
		got := lb.Next()
		if got == nil {
			t.Fatalf("第 %d 次调用 Next() 返回 nil", i)
		}
		if got.Host != want {
			t.Errorf("第 %d 次调用 Next() 返回 Host = %s，期望 %s", i, got.Host, want)
		}
	}
}

// 测试 UpdateBackends：动态替换后端列表后，轮询基于新列表
func TestLoadBalancerUpdateBackends(t *testing.T) {
	initial := []*url.URL{{Scheme: "http", Host: "old1:8080"}, {Scheme: "http", Host: "old2:8080"}}
	lb := New(initial)

	// 先取几次，验证初始列表
	first := lb.Next()
	if first.Host != "old1:8080" {
		t.Errorf("初始第一次应为 old1:8080，得到 %s", first.Host)
	}

	newBackends := []*url.URL{{Scheme: "http", Host: "new1:9090"}, {Scheme: "http", Host: "new2:9090"}}
	lb.UpdateBackends(newBackends)

	// 更新后，Next 应该从新列表的第一个开始（注意：counter 没有重置，但取模后依然正确）
	got1 := lb.Next()
	got2 := lb.Next()
	if got1.Host != "new1:9090" {
		t.Errorf("更新后第一次应为 new1:9090，得到 %s", got1.Host)
	}
	if got2.Host != "new2:9090" {
		t.Errorf("更新后第二次应为 new2:9090，得到 %s", got2.Host)
	}
}

// 测试并发安全性：多个 goroutine 同时调用 Next()，不应出现数据竞争
func TestLoadBalancerConcurrent(t *testing.T) {
	backends := testBackends()
	lb := New(backends)

	const goroutines = 100
	const callsPerGoroutine = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < callsPerGoroutine; j++ {
				_ = lb.Next()
			}
		}()
	}
	wg.Wait()
	// 如果没有触发 race condition，测试即通过（用 go test -race 更严格）
}

// 基准测试：测量 Next() 的单次性能
func BenchmarkLoadBalancerNext(b *testing.B) {
	backends := testBackends()
	lb := New(backends)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = lb.Next()
	}
}
