package log

import (
	"io" // 用于多输出目标（控制台 + 文件）
	"log/slog"
	"os"
	"strings" // 用于解析日志级别字符串
)

// / 根据配置初始化全局 slog
// / # 参数
// / - level: 日志级别字符串 (debug, info, warn, error)
// / - filePath: 输出文件路径，为空则只输出到控制台
// / # 返回
// / - error: 初始化错误（如果有）
func Init(level string, filePath string) error {
	// 解析日志级别
	var logLevel slog.Level

	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo // 默认级别
	}

	// 构建输出目标
	var writers []io.Writer
	writers = append(writers, os.Stderr) // 始终输出到控制台

	if filePath != "" {
		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

		if err != nil {
			return err
		}

		writers = append(writers, file)
	}

	// 多输出目标：控制台 + 文件（如果配置了）
	multiWriter := io.MultiWriter(writers...)

	// 创建 JSON Handler，并设置日志级别过滤
	handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: logLevel,
	})

	// 设置全局默认 Logger，所有通过 slog 包记录的日志都会使用这个 Handler 和级别过滤
	slog.SetDefault(slog.New(handler))

	return nil
}
