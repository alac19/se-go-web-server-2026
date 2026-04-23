package main

import (
	// 导入标准库
	"context"
	"crypto/tls" // 用于 TLS 配置
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os" // 用于退出程序

	// 导入本地包
	"se-go-web-server-2026/src/config"
	"se-go-web-server-2026/src/health"
	"se-go-web-server-2026/src/lb"
	"se-go-web-server-2026/src/proxy"
)

// 程序入口，负责加载配置、初始化组件（负载均衡器、健康检查器、反向代理）并启动 HTTP/HTTPS 服务器
func main() {
	fmt.Println("Hello, Cloud Native! 项目进入代码编写阶段！")

	// 加载配置
	cfg, err := config.LoadConfig("configs/config.toml")

	if err != nil {
		slog.Error("加载配置失败", "error", err)
		os.Exit(1)
	}

	if len(cfg.Listen) < 2 {
		slog.Error("配置文件中至少需要两个监听地址（HTTP + HTTPS）")
		os.Exit(1)
	}

	// 创建负载均衡器（初始为空）
	lbInstance := lb.New([]*url.URL{})

	// 准备所有后端 URL（用于健康检查）
	allURLs := make([]*url.URL, len(cfg.Forward))

	for i, f := range cfg.Forward {
		allURLs[i] = &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", f.ForwardAddr, f.ForwardPort),
		}
	}

	// 健康检查回调：更新负载均衡器的健康后端列表
	onHealthy := func(healthy []*url.URL) {
		lbInstance.UpdateBackends(healthy)
		slog.Info("健康后端更新", "count", len(healthy))
	}

	// 启动健康检查（如果启用）
	if cfg.Heartbeat.Enabled {
		checker := health.NewChecker(
			allURLs,
			cfg.Heartbeat.Path,
			cfg.Heartbeat.IntervalSeconds,
			cfg.Heartbeat.TimeoutSeconds,
			onHealthy,
		)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		checker.Start(ctx)
		slog.Info("健康检查已启动", "interval", cfg.Heartbeat.IntervalSeconds, "path", cfg.Heartbeat.Path)
	} else {
		// 未启用健康检查：将所有后端视为健康
		lbInstance.UpdateBackends(allURLs)
		slog.Warn("健康检查未启用，所有后端均被视为健康")
	}

	// 创建反向代理处理器
	// 注意：cfg.Proxy.TimeoutSeconds 可能为零值（如果配置中没有 [proxy] 段），此时默认超时设为 30 秒
	timeoutSec := cfg.Proxy.TimeoutSeconds

	if timeoutSec <= 0 {
		timeoutSec = 30
		slog.Warn("未配置代理超时，使用默认值 30 秒")
	}

	proxyHandler := proxy.NewReverseProxy(lbInstance, timeoutSec)

	// 启动 HTTP 服务器（第一个监听）
	httpAddr := fmt.Sprintf("%s:%d", cfg.Listen[0].ListenAddr, cfg.Listen[0].ListenPort)
	startHTTPServer(httpAddr, proxyHandler)

	// 启动 HTTPS 服务器（第二个监听）
	httpsAddr := fmt.Sprintf("%s:%d", cfg.Listen[1].ListenAddr, cfg.Listen[1].ListenPort)
	startHTTPSServer(httpsAddr, proxyHandler, cfg.TLS)

	// 阻塞主协程（保持运行）
	select {}
}

// 启动 HTTP 服务器（非 TLS）
// / # 参数
// / - addr: 监听地址（IP:端口）
// / - handler: HTTP 处理器（反向代理）
func startHTTPServer(addr string, handler http.Handler) {
	// 因为 ListenAndServe 是阻塞的，所以在一个新的协程中启动服务器
	go func() {
		slog.Info("HTTP 服务器启动", "addr", addr)

		if err := http.ListenAndServe(addr, handler); err != nil {
			slog.Error("HTTP 服务器错误", "error", err)
		}
	}()
}

// 启动 HTTPS 服务器（TLS）
// / # 参数
// / - addr: 监听地址（IP:端口）
// / - handler: HTTP 处理器（反向代理）
func startHTTPSServer(addr string, handler http.Handler, tlsCfg config.TLSConfig) {
	// 加载 TLS 证书
	// 如果配置中没有指定证书路径，使用默认值
	certFile := tlsCfg.CertFile

	if certFile == "" {
		certFile = "cert.pem"
		slog.Warn("未配置 cert_file，使用默认值 cert.pem")
	}

	keyFile := tlsCfg.KeyFile

	if keyFile == "" {
		keyFile = "key.pem"
		slog.Warn("未配置 key_file，使用默认值 key.pem")
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)

	if err != nil {
		slog.Error("加载 TLS 证书失败", "error", err, "cert", certFile, "key", keyFile)
		return
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,           // 强制使用 TLS 1.2 或更高版本
		NextProtos:   []string{"h2", "http/1.1"}, // 设置协议版本，支持 HTTP/2
	}
	server := &http.Server{
		Addr:      addr,
		Handler:   handler,
		TLSConfig: tlsConfig,
	}

	go func() {
		slog.Info("HTTPS 服务器启动", "addr", addr)

		// 因为已经设置了 TLSConfig，传入空字符串表示使用配置中的证书，ListenAndServeTLS 内部会调用 tlsConfig 来获取证书
		if err := server.ListenAndServeTLS("", ""); err != nil {
			slog.Error("HTTPS 服务器错误", "error", err)
		}
	}()
}
