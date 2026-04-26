package config

import (
	"os"
	"path/filepath"
	"testing"
)

// 测试正常加载配置
func TestLoadConfigSuccess(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	content := `
[[listen]]
listen_addr = "127.0.0.1"
listen_port = 8080

[[listen]]
listen_addr = "127.0.0.1"
listen_port = 8443

[[forward]]
forward_addr = "127.0.0.1"
forward_port = 9091

[[forward]]
forward_addr = "127.0.0.1"
forward_port = 9092
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig 返回错误: %v", err)
	}
	if cfg == nil {
		t.Fatal("cfg 为 nil")
	}
	if len(cfg.Listen) != 2 {
		t.Errorf("期望 2 个 listen，实际 %d", len(cfg.Listen))
	}
	if len(cfg.Forward) != 2 {
		t.Errorf("期望 2 个 forward，实际 %d", len(cfg.Forward))
	}
	if cfg.Listen[0].ListenPort != 8080 {
		t.Errorf("第一个监听端口应为 8080，实际 %d", cfg.Listen[0].ListenPort)
	}
}

// 测试文件不存在
func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := LoadConfig("/no/such/file.toml")
	if err == nil {
		t.Fatal("期望错误，但没有返回错误")
	}
	// 可选：检查错误信息是否包含 "decode config failed"
}

// 测试 TOML 格式错误
func TestLoadConfigInvalidToml(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.toml")
	content := `this is not valid toml = [`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("写入文件失败: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Fatal("期望解析错误，但没有返回错误")
	}
}

// 测试缺少 listen 项
func TestLoadConfigNoListen(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "no_listen.toml")
	content := `
[[forward]]
forward_addr = "127.0.0.1"
forward_port = 9091
`
	os.WriteFile(configPath, []byte(content), 0644)

	_, err := LoadConfig(configPath)
	if err == nil || err.Error() != "配置中没有 listen 项" {
		t.Errorf("期望 '配置中没有 listen 项' 错误，得到 %v", err)
	}
}

// 测试缺少 forward 项
func TestLoadConfigNoForward(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "no_forward.toml")
	content := `
[[listen]]
listen_addr = "127.0.0.1"
listen_port = 8080
`
	os.WriteFile(configPath, []byte(content), 0644)

	_, err := LoadConfig(configPath)
	if err == nil || err.Error() != "配置中没有 forward 项" {
		t.Errorf("期望 '配置中没有 forward 项' 错误，得到 %v", err)
	}
}

// 测试包含心跳配置的完整配置
func TestLoadConfigWithHeartbeat(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	content := `
[[listen]]
listen_addr = "127.0.0.1"
listen_port = 8080

[[forward]]
forward_addr = "127.0.0.1"
forward_port = 9091

[heartbeat]
enabled = true
path = "/ping"
interval_seconds = 10
timeout_seconds = 2
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig 返回错误: %v", err)
	}
	if !cfg.Heartbeat.Enabled {
		t.Error("Heartbeat.Enabled 应为 true")
	}
	if cfg.Heartbeat.Path != "/ping" {
		t.Errorf("Heartbeat.Path 期望 /ping，得到 %s", cfg.Heartbeat.Path)
	}
	if cfg.Heartbeat.IntervalSeconds != 10 {
		t.Errorf("IntervalSeconds 期望 10，得到 %d", cfg.Heartbeat.IntervalSeconds)
	}
	if cfg.Heartbeat.TimeoutSeconds != 2 {
		t.Errorf("TimeoutSeconds 期望 2，得到 %d", cfg.Heartbeat.TimeoutSeconds)
	}
}

// 测试包含 proxy 超时配置的完整配置
func TestLoadConfigWithProxy(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	content := `
[[listen]]
listen_addr = "127.0.0.1"
listen_port = 8080

[[forward]]
forward_addr = "127.0.0.1"
forward_port = 9091

[heartbeat]
enabled = true
path = "/heartbeat"
interval_seconds = 30
timeout_seconds = 2

[proxy]
timeout_seconds = 45
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig 返回错误: %v", err)
	}
	if cfg.Proxy.TimeoutSeconds != 45 {
		t.Errorf("Proxy.TimeoutSeconds 期望 45，得到 %d", cfg.Proxy.TimeoutSeconds)
	}
}

// 测试缺少 proxy 配置时，TimeoutSeconds 应为零值（0）
func TestLoadConfigMissingProxy(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	content := `
[[listen]]
listen_addr = "127.0.0.1"
listen_port = 8080

[[forward]]
forward_addr = "127.0.0.1"
forward_port = 9091
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig 返回错误: %v", err)
	}
	if cfg.Proxy.TimeoutSeconds != 0 {
		t.Errorf("缺少 proxy 段时 TimeoutSeconds 应为 0，得到 %d", cfg.Proxy.TimeoutSeconds)
	}
}

// 测试包含 TLS 证书配置的完整配置
func TestLoadConfigWithTLS(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	content := `
[[listen]]
listen_addr = "127.0.0.1"
listen_port = 8080

[[forward]]
forward_addr = "127.0.0.1"
forward_port = 9091

[heartbeat]
enabled = true
path = "/heartbeat"
interval_seconds = 30
timeout_seconds = 2

[proxy]
timeout_seconds = 45

[tls]
cert_file = "mycert.pem"
key_file = "mykey.pem"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig 返回错误: %v", err)
	}
	if cfg.TLS.CertFile != "mycert.pem" {
		t.Errorf("TLS.CertFile 期望 mycert.pem，得到 %s", cfg.TLS.CertFile)
	}
	if cfg.TLS.KeyFile != "mykey.pem" {
		t.Errorf("TLS.KeyFile 期望 mykey.pem，得到 %s", cfg.TLS.KeyFile)
	}
}

// 测试缺少 TLS 配置时，CertFile 和 KeyFile 应为空字符串
func TestLoadConfigMissingTLS(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	content := `
[[listen]]
listen_addr = "127.0.0.1"
listen_port = 8080

[[forward]]
forward_addr = "127.0.0.1"
forward_port = 9091
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig 返回错误: %v", err)
	}
	if cfg.TLS.CertFile != "" {
		t.Errorf("缺少 TLS 段时 CertFile 应为空，得到 %s", cfg.TLS.CertFile)
	}
	if cfg.TLS.KeyFile != "" {
		t.Errorf("缺少 TLS 段时 KeyFile 应为空，得到 %s", cfg.TLS.KeyFile)
	}
}

// 测试包含日志配置的完整配置
func TestLoadConfigWithLog(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	content := `
[[listen]]
listen_addr = "127.0.0.1"
listen_port = 8080

[[forward]]
forward_addr = "127.0.0.1"
forward_port = 9091

[heartbeat]
enabled = true
path = "/heartbeat"
interval_seconds = 30
timeout_seconds = 2

[proxy]
timeout_seconds = 45

[tls]
cert_file = "mycert.pem"
key_file = "mykey.pem"

[log]
level = "debug"
file_path = "myapp.log"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig 返回错误: %v", err)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("Log.Level 期望 debug，得到 %s", cfg.Log.Level)
	}
	if cfg.Log.FilePath != "myapp.log" {
		t.Errorf("Log.FilePath 期望 myapp.log，得到 %s", cfg.Log.FilePath)
	}
}

// 测试缺少日志配置时，Level 和 FilePath 应为空字符串
func TestLoadConfigMissingLog(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	content := `
[[listen]]
listen_addr = "127.0.0.1"
listen_port = 8080

[[forward]]
forward_addr = "127.0.0.1"
forward_port = 9091
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig 返回错误: %v", err)
	}
	if cfg.Log.Level != "" {
		t.Errorf("缺少 log 段时 Level 应为空，得到 %s", cfg.Log.Level)
	}
	if cfg.Log.FilePath != "" {
		t.Errorf("缺少 log 段时 FilePath 应为空，得到 %s", cfg.Log.FilePath)
	}
}

// 测试心跳配置校验：间隔为0
func TestLoadConfigHeartbeatZeroInterval(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	content := `
[[listen]]
listen_addr = "127.0.0.1"
listen_port = 8080

[[forward]]
forward_addr = "127.0.0.1"
forward_port = 9091

[heartbeat]
enabled = true
path = "/heartbeat"
interval_seconds = 0
timeout_seconds = 2
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}
	_, err := LoadConfig(configPath)
	if err == nil {
		t.Fatal("期望心跳间隔为0时返回错误，但没有错误")
	}
	if err.Error() != "心跳间隔 interval_seconds 必须为正数" {
		t.Errorf("期望错误信息 '心跳间隔 interval_seconds 必须为正数'，得到 %v", err)
	}
}

// 测试心跳配置校验：超时为0
func TestLoadConfigHeartbeatZeroTimeout(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	content := `
[[listen]]
listen_addr = "127.0.0.1"
listen_port = 8080

[[forward]]
forward_addr = "127.0.0.1"
forward_port = 9091

[heartbeat]
enabled = true
path = "/heartbeat"
interval_seconds = 30
timeout_seconds = 0
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}
	_, err := LoadConfig(configPath)
	if err == nil {
		t.Fatal("期望心跳超时为0时返回错误，但没有错误")
	}
	if err.Error() != "心跳超时 timeout_seconds 必须为正数" {
		t.Errorf("期望错误信息 '心跳超时 timeout_seconds 必须为正数'，得到 %v", err)
	}
}

// 测试心跳配置校验：路径为空
func TestLoadConfigHeartbeatEmptyPath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	content := `
[[listen]]
listen_addr = "127.0.0.1"
listen_port = 8080

[[forward]]
forward_addr = "127.0.0.1"
forward_port = 9091

[heartbeat]
enabled = true
path = ""
interval_seconds = 30
timeout_seconds = 2
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}
	_, err := LoadConfig(configPath)
	if err == nil {
		t.Fatal("期望心跳路径为空时返回错误，但没有错误")
	}
	if err.Error() != "心跳路径 path 不能为空" {
		t.Errorf("期望错误信息 '心跳路径 path 不能为空'，得到 %v", err)
	}
}
