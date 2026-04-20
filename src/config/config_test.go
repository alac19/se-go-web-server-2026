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
