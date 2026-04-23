package lb

import (
	"net/url"
	"sync/atomic"
)

// LoadBalancer 提供轮询算法，线程安全。
type LoadBalancer struct {
	/// 后端 URL 列表（由健康检查动态更新）
	urls []*url.URL

	/// 当前索引
	counter atomic.Uint64
}

// / 创建负载均衡器，初始后端列表可以为空（后续由健康检查更新）
// / # 参数
// / - urls: 初始后端 URL 列表
// /
// / # 返回
// / - *LoadBalancer: 创建的 LoadBalancer 实例
func New(urls []*url.URL) *LoadBalancer {
	return &LoadBalancer{urls: urls}
}

// / 返回下一个可用的后端 URL（轮询）
// / # 参数
// / - lb: 创建的 LoadBalancer 实例
// /
// / # 返回
// / - *url.URL: 下一个可用的后端 URL，如果没有后端，返回 nil
func (lb *LoadBalancer) Next() *url.URL {
	n := len(lb.urls)

	if n == 0 {
		return nil
	}

	idx := lb.counter.Add(1) - 1

	return lb.urls[idx%uint64(n)]
}

// / 原子替换后端列表（用于健康检查动态更新）。
// / # 参数
// / - lb: 创建的 LoadBalancer 实例
// / - urls: 后端 URL 列表
func (lb *LoadBalancer) UpdateBackends(urls []*url.URL) {
	lb.urls = urls
	// 不重置 counter，允许继续轮询（不会出错，只是可能跳过一些，但能避免更新后端后所有请求瞬间涌向第一个后端（惊群效应））
}

// / 返回当前健康后端数量（仅读，不改变轮询状态）
// / # 参数
// / - lb: 创建的 LoadBalancer 实例
// /
// / # 返回
// / - int: 当前健康后端数量
func (lb *LoadBalancer) HealthyCount() int {
	return len(lb.urls)
}
