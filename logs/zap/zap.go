package zap

import (
	"errors"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

// Logger 避免业务代码对zap直接依赖
type Logger = zap.SugaredLogger

// NewLogger 创建zap
func NewLogger(conf Config) *Logger {
	return zap.New(newZapCore(conf), zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
}

func newZapCore(conf Config) zapcore.Core {
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(parseLevel(conf.Level))

	writers := make([]zapcore.WriteSyncer, 0)
	if strings.Contains(conf.Writer, "console") {
		writers = append(writers, zapcore.AddSync(os.Stdout))
	}
	if strings.Contains(conf.Writer, "file") {
		writers = append(writers, zapcore.AddSync(&lumberjack.Logger{
			Filename:   conf.FileConfig.Path,       // 日志文件路径
			Compress:   conf.FileConfig.Compress,   // 是否压缩
			MaxSize:    conf.FileConfig.MaxSize,    // 每个日志文件保存的最大尺寸 单位：M
			MaxAge:     conf.FileConfig.MaxAge,     // 文件最多保存多少天
			MaxBackups: conf.FileConfig.MaxBackups, // 日志文件最多保存多少个备份
		}))
	}
	if len(writers) == 0 {
		panic(errors.New("logs writer not set: console or file"))
	}

	return zapcore.NewCore(
		zapcore.NewJSONEncoder(defaultEncoderConfig), // 日志格式
		zapcore.NewMultiWriteSyncer(writers...),      // 打印到控制台和文件
		atomicLevel,                                  // 日志级别
	)
}

var defaultEncoderConfig = zapcore.EncoderConfig{
	CallerKey:      "caller",
	StacktraceKey:  "stack",
	TimeKey:        "time",
	MessageKey:     "msg",
	LevelKey:       "level",
	NameKey:        "logger",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeCaller:   zapcore.ShortCallerEncoder,
	EncodeLevel:    zapcore.CapitalLevelEncoder,
	EncodeDuration: zapcore.StringDurationEncoder,
	EncodeName:     zapcore.FullNameEncoder,
	EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	},
}

func parseLevel(level string) zapcore.Level {
	switch level {
	case "fatal":
		return zap.FatalLevel
	case "panic":
		return zap.PanicLevel
	case "error":
		return zap.ErrorLevel
	case "warn", "warning":
		return zap.WarnLevel
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	}
	return zap.InfoLevel
}
