package log

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInit_LevelParsing(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"debug", false},
		{"info", false},
		{"warn", false},
		{"error", false},
		{"DEBUG", false},
		{"Info", false},
		{"invalid", false}, // 默认 info
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			err := Init(tt.input, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInit_FileOutput(t *testing.T) {
	// 创建临时目录（手动管理，避免自动删除失败）
	tmpDir, err := os.MkdirTemp("", "logtest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir) // 测试结束后删除，如果失败将忽略

	logPath := filepath.Join(tmpDir, "test.log")
	err = Init("info", logPath)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	// 检查文件是否创建
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("日志文件未创建")
	}

	// 测试无效路径
	invalidPath := filepath.Join(tmpDir, "nonexist", "log.txt")
	err = Init("info", invalidPath)
	if err == nil {
		t.Error("期望因目录不存在而返回错误，但没有错误")
	}
}

func TestInit_DefaultStderr(t *testing.T) {
	err := Init("info", "")
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	// 没有错误即成功
}
