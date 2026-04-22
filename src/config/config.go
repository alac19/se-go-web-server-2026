// src/config/config.go
package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// Config 总配置
type Config struct {
	Listen    []Listen  `toml:"listen"`
	Forward   []Forward `toml:"forward"`
	Heartbeat Heartbeat `toml:"heartbeat"`
	Proxy     Proxy     `toml:"proxy"`
}

// Listen 监听配置
type Listen struct {
	ListenAddr string `toml:"listen_addr"` // 结构体标签，指定 TOML 中的字段名
	ListenPort uint16 `toml:"listen_port"` // 适配端口号（范围 0~65535）
}

// Forward 后端转发配置
type Forward struct {
	ForwardAddr string `toml:"forward_addr"`
	ForwardPort uint16 `toml:"forward_port"`
}

// Heartbeat 心跳检测配置
type Heartbeat struct {
	Enabled         bool   `toml:"enabled"`          // 是否启用
	Path            string `toml:"path"`             // 健康检查路径
	IntervalSeconds int    `toml:"interval_seconds"` // 检查间隔（秒）
	TimeoutSeconds  int    `toml:"timeout_seconds"`  // 单次请求超时（秒）
}

// Proxy 代理转发配置
type Proxy struct {
	TimeoutSeconds int `toml:"timeout_seconds"` // 转发请求超时（秒）
}

// / 从指定路径加载 TOML 配置文件
// / # 参数
// / - path: 配置文件路径
// /
// / # 返回
// / - *Config: 配置对象指针
// / - error: 错误信息
func LoadConfig(path string) (*Config, error) {
	var config Config

	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, fmt.Errorf("decode config failed: %w", err)
	}

	// 校验
	if len(config.Listen) == 0 {
		return nil, fmt.Errorf("配置中没有 listen 项")
	}
	if len(config.Forward) == 0 {
		return nil, fmt.Errorf("配置中没有 forward 项")
	}

	return &config, nil
}
