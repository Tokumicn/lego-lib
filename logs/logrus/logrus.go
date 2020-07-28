package logrus

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// FileConfig 日志配置项 文件类型时需要配置
type FileConfig struct {
	Path       string `toml:"path"`
	Compress   bool   `toml:"compress"`
	MaxSize    int    `toml:"max_size"`
	MaxAge     int    `toml:"max_age"`
	MaxBackups int    `toml:"max_backups"`
}

// Config 日志配置项
type Config struct {
	Writer     string     `toml:"writer"`
	Level      string     `toml:"level"`
	FileConfig FileConfig `toml:"file_config"`
}

// Logger 避免业务代码对logrus直接依赖
type Logger = logrus.Logger

// NewLogger 创建logrus
func NewLogger(conf Config) *Logger {
	logger := logrus.New()

	level, err := logrus.ParseLevel(conf.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
	})

	writers := make([]io.Writer, 0)
	if strings.Contains(conf.Writer, "console") {
		writers = append(writers, os.Stdout)
	}
	if strings.Contains(conf.Writer, "file") {
		writers = append(writers, &lumberjack.Logger{
			Filename:   conf.FileConfig.Path,       // 日志文件路径
			Compress:   conf.FileConfig.Compress,   // 是否压缩
			MaxSize:    conf.FileConfig.MaxSize,    // 每个日志文件保存的最大尺寸 单位：M
			MaxAge:     conf.FileConfig.MaxAge,     // 文件最多保存多少天
			MaxBackups: conf.FileConfig.MaxBackups, // 日志文件最多保存多少个备份
		})
	}
	if len(writers) == 0 {
		panic(errors.New("logs writer not set: console or file"))
	}

	logger.SetOutput(io.MultiWriter(writers...))
	return logger
}
