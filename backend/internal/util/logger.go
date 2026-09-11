// Package util 提供日志、JWT、错误、格式化等通用工具。
package util

import (
	"log/slog"
	"os"
	"strings"
)

// NewLogger 创建结构化 slog 日志器。
func NewLogger(mode string) *slog.Logger {
	level := slog.LevelInfo
	if strings.EqualFold(mode, "debug") {
		level = slog.LevelDebug
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
