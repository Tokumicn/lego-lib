package logs

import (
	"github.com/Tokumicn/lego-lib/logs/logrus"
	"github.com/Tokumicn/lego-lib/logs/zap"
)

// FileConfig 日志配置项 文件类型时需要配置
type FileConfig struct {
	Path       string `toml:"path"`
	Compress   bool   `toml:"compress"`
	MaxSize    int    `toml:"max_size"`
	MaxAge     int    `toml:"max_age"`
	MaxBackups int    `toml:"max_backups"`
}

// Config 日志配置项 Writer 可配置为 file; console; file,console
type Config struct {
	Writer     string     `toml:"writer"`
	Level      string     `toml:"level"`
	FileConfig FileConfig `toml:"file_config"`
}

// Logger 日志接口
type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	Fatal(v ...interface{})

	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
}

var (
	defaultLogger     Logger
	defaultDataLogger Logger
)

// Init Log全局初始化函数 调用方保证全局调用一次
func Init(conf *Config) {
	defaultLogger = initZap(conf)
	//defaultLogger = initLogrus(conf)
}

// InitData 数据Log全局初始化函数 调用方保证全局调用一次
func InitData(conf *Config) {
	defaultLogger = initZap(conf)
	//defaultLogger = initLogrus(conf)
}

func initZap(conf *Config) *zap.Logger {
	return zap.NewLogger(zap.Config{
		Writer: conf.Writer,
		FileConfig: zap.FileConfig{
			Path:       conf.FileConfig.Path,
			Compress:   conf.FileConfig.Compress,
			MaxSize:    conf.FileConfig.MaxSize,
			MaxAge:     conf.FileConfig.MaxAge,
			MaxBackups: conf.FileConfig.MaxBackups,
		},
		Level: conf.Level,
	})
}

func initLogrus(conf *Config) *logrus.Logger {
	return logrus.NewLogger(logrus.Config{
		Writer: conf.Writer,
		FileConfig: logrus.FileConfig{
			Path:       conf.FileConfig.Path,
			Compress:   conf.FileConfig.Compress,
			MaxSize:    conf.FileConfig.MaxSize,
			MaxAge:     conf.FileConfig.MaxAge,
			MaxBackups: conf.FileConfig.MaxBackups,
		},
		Level: conf.Level,
	})
}

// Debug 打印Debug日志
func Debug(v ...interface{}) {
	defaultLogger.Debug(v...)
}

// Info 打印Info日志
func Info(v ...interface{}) {
	defaultLogger.Info(v...)
}

// Warn 打印Warn日志
func Warn(v ...interface{}) {
	defaultLogger.Warn(v...)
}

// Error 打印Error日志
func Error(v ...interface{}) {
	defaultLogger.Error(v...)
}

// Fatal 打印Fatal日志
func Fatal(v ...interface{}) {
	defaultLogger.Fatal(v...)
}

// Debugf 打印Debug日志
func Debugf(format string, v ...interface{}) {
	defaultLogger.Debugf(format, v...)
}

// Infof 打印Info日志
func Infof(format string, v ...interface{}) {
	defaultLogger.Infof(format, v...)
}

// Warnf 打印Warn日志
func Warnf(format string, v ...interface{}) {
	defaultLogger.Warnf(format, v...)
}

// Errorf 打印Error日志
func Errorf(format string, v ...interface{}) {
	defaultLogger.Errorf(format, v...)
}

// Fatalf 打印Fatal日志
func Fatalf(format string, v ...interface{}) {
	defaultLogger.Fatalf(format, v...)
}

// Dataf 打印数据日志
func Dataf(format string, v ...interface{}) {
	defaultDataLogger.Infof(format, v...)
}
