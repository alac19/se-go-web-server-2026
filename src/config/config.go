// src/config/config.go
package config

import (
	"fmt"

	"github.com/BurntSushi/toml" // 用于解析 TOML 配置文件
)

// Config 总配置
type Config struct {
	Listen    []Listen  `toml:"listen"`
	Forward   []Forward `toml:"forward"`
	Heartbeat Heartbeat `toml:"heartbeat"`
	Proxy     Proxy     `toml:"proxy"`
	TLS       TLSConfig `toml:"tls"`
	Log       Log       `toml:"log"`
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

// TLS 配置
type TLSConfig struct {
	CertFile string `toml:"cert_file"`
	KeyFile  string `toml:"key_file"`
}

// Log 日志配置
type Log struct {
	Level    string `toml:"level"`     // 日志级别，debug, info, warn, error
	FilePath string `toml:"file_path"` // 日志文件路径，空值表示输出到控制台
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
	for _, l := range config.Listen {
		if l.ListenPort == 0 {
			return nil, fmt.Errorf("监听端口不能为 0")
		}
	}
	for _, f := range config.Forward {
		if f.ForwardPort == 0 {
			return nil, fmt.Errorf("转发端口不能为 0")
		}
	}
	if config.Heartbeat.Enabled {
		if config.Heartbeat.IntervalSeconds <= 0 {
			return nil, fmt.Errorf("心跳间隔 interval_seconds 必须为正数")
		}
		if config.Heartbeat.TimeoutSeconds <= 0 {
			return nil, fmt.Errorf("心跳超时 timeout_seconds 必须为正数")
		}
		if config.Heartbeat.Path == "" {
			return nil, fmt.Errorf("心跳路径 path 不能为空")
		}
	}

	return &config, nil
}
