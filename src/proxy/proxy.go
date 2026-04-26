package proxy

import (
	"log/slog"
	"net/http"
	"net/http/httputil" // 用于实现反向代理
	"se-go-web-server-2026/src/lb"
	"time"
)

// / 创建反向代理，使用给定的负载均衡器选择后端
// / # 参数
// / - lb: 负载均衡器实例，用于选择后端 URL
// / - timeoutSec: 转发请求的超时时间（秒）
// /
// / # 返回
// / - http.Handler: 处理 HTTP 请求的 Handler，内部实现反向代理逻辑
func NewReverseProxy(lb *lb.LoadBalancer, timeoutSec int) http.Handler {
	// 配置 HTTP 传输层（连接池、超时等）
	transport := &http.Transport{
		MaxIdleConns:          300,                                     // 全局最大空闲连接数
		MaxIdleConnsPerHost:   100,                                     // 每个后端的最大空闲连接数
		IdleConnTimeout:       90 * time.Second,                        // 空闲连接超时
		TLSHandshakeTimeout:   10 * time.Second,                        // TLS 握手超时
		ResponseHeaderTimeout: time.Duration(timeoutSec) * time.Second, // 整体超时
		// 显式启用 Keep-Alive（默认就是true，但确保）
		DisableKeepAlives: false,
	}
	// 定义反向代理的请求转发器（Director）
	// 当有健康后端时，这个 director 会被调用，修改请求的 URL 为目标后端
	director := func(req *http.Request) {
		backend := lb.Next()

		if backend == nil {
			return
		}

		// 修改请求的 URL 为目标后端
		req.URL.Scheme = backend.Scheme
		req.URL.Host = backend.Host
		// 保留原始路径和查询参数（ReverseProxy 默认会保留）
	}

	// 创建实际的反向代理对象
	rp := &httputil.ReverseProxy{ // httputil.ReverseProxy 通过 ServeHTTP 方法自动构造转发请求、复制请求头（过滤逐跳头部）、流式转发请求体、接收后端响应、流式返回给客户端
		Director:  director,
		Transport: transport,
		ErrorHandler: func(rw http.ResponseWriter, req *http.Request, err error) {
			// 当转发过程中出现错误（如后端连接失败、超时），记录错误并返回 502
			slog.Error("反向代理转发失败", "method", req.Method, "path", req.URL.Path, "error", err)
			http.Error(rw, "Bad Gateway", http.StatusBadGateway)
		},
	}

	// 返回一个 Handler 闭包函数（装饰器模式），外部调用时会执行这个函数，它先检查健康后端数量，再决定是返回 503 还是调用反向代理
	// w：http.ResponseWriter 接口，用于构造 HTTP 响应（设置状态码、头部、写入响应体）。
	// r：*http.Request 指针，代表客户端发来的 HTTP 请求（包含方法、URL、头部、请求体等）。
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 如果没有健康后端，直接返回 503
		if lb.HealthyCount() == 0 {
			slog.Warn("所有后端均不可用", "path", r.URL.Path)
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}

		// 有健康后端，交给反向代理处理
		rp.ServeHTTP(w, r) // ServeHTTP 是 http.Handler 接口的核心方法。任何实现了该接口的类型（比如 httputil.ReverseProxy、http.HandlerFunc 等）都可以通过调用 ServeHTTP(w, r) 来处理请求，它会把请求和响应交给真正的反向代理去处理
	})
}
