package proxy

import (
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
// / - *httputil.ReverseProxy: 创建的反向代理实例
func NewReverseProxy(lb *lb.LoadBalancer, timeoutSec int) *httputil.ReverseProxy { // httputil.ReverseProxy 自动构造转发请求、复制请求头（过滤逐跳头部）、流式转发请求体、接收后端响应、流式返回给客户端
	transport := &http.Transport{
		MaxIdleConns:          100,                                     // 全局最大空闲连接数
		MaxIdleConnsPerHost:   20,                                      // 每个后端的最大空闲连接数
		IdleConnTimeout:       90 * time.Second,                        // 空闲连接超时
		TLSHandshakeTimeout:   10 * time.Second,                        // TLS 握手超时
		ResponseHeaderTimeout: time.Duration(timeoutSec) * time.Second, // 整体超时
	}
	director := func(req *http.Request) {
		backend := lb.Next()

		if backend == nil {
			// 没有可用后端，将在 ErrorHandler 中返回 502
			return
		}

		// 修改请求的 URL 为目标后端
		req.URL.Scheme = backend.Scheme
		req.URL.Host = backend.Host
		// 保留原始路径和查询参数（ReverseProxy 默认会保留）
	}

	return &httputil.ReverseProxy{
		Director:  director,
		Transport: transport,
		ErrorHandler: func(rw http.ResponseWriter, req *http.Request, err error) {
			http.Error(rw, "Bad Gateway", http.StatusBadGateway)
		},
	}
}
