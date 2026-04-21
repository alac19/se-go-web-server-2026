package main // 声明这是一个可执行程序

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"se-go-web-server-2026/src/config"
	"se-go-web-server-2026/src/health"
	"se-go-web-server-2026/src/lb"
)

func main() {
	fmt.Println("Hello, Cloud Native! 项目进入代码编写阶段！") // 输出一行文本到控制台

	// 1. 加载配置
	cfg, _ := config.LoadConfig("configs/config.toml")

	// 2. 创建负载均衡器
	lb := lb.New([]*url.URL{})

	// 3. 创建健康检查器并启动（暂时注释掉代理部分）
	onHealthy := func(healthy []*url.URL) {
		lb.UpdateBackends(healthy)
		slog.Info("健康后端更新", "count", len(healthy))
	}
	allURLs := make([]*url.URL, len(cfg.Forward))
	for i, f := range cfg.Forward {
		allURLs[i] = &url.URL{Scheme: "http", Host: fmt.Sprintf("%s:%d", f.ForwardAddr, f.ForwardPort)}
	}
	checker := health.NewChecker(allURLs, cfg.Heartbeat.Path, cfg.Heartbeat.IntervalSeconds, cfg.Heartbeat.TimeoutSeconds, onHealthy)
	checker.Start(context.Background())

	// 4. 阻塞住，不然 main 会退出
	select {}
}
