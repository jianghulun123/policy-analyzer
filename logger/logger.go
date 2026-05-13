// Package logger 提供结构化日志功能
//
// 日志级别：DEBUG < INFO < WARN < ERROR
// 输出格式：时间戳 | 级别 | 模块 | 消息 | 字段...
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Level 日志级别
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

var levelNames = map[Level]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
}

var levelColors = map[Level]string{
	LevelDebug: "\033[36m", // Cyan
	LevelInfo:  "\033[32m", // Green
	LevelWarn:  "\033[33m", // Yellow
	LevelError: "\033[31m", // Red
}

const colorReset = "\033[0m"

// Field 日志字段
type Field struct {
	Key   string
	Value interface{}
}

// Logger 日志记录器
type Logger struct {
	module     string
	level      Level
	writer     io.Writer
	fileWriter io.Writer
	mu         sync.Mutex
	useColor   bool
	jsonFormat bool
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Config 日志配置
type Config struct {
	Level      Level
	Module     string
	LogDir     string // 日志文件目录，为空则不写文件
	UseColor   bool
	JSONFormat bool
}

// Init 初始化默认日志记录器
func Init(cfg Config) {
	once.Do(func() {
		defaultLogger = NewLogger(cfg)
		// 更新全局记录器引用，让已创建的模块记录器也能使用文件写入
		globalFileWriter = defaultLogger.fileWriter
	})
}

// globalFileWriter 全局文件写入器，供子记录器使用
var globalFileWriter io.Writer

// GetLogger 获取默认日志记录器
func GetLogger() *Logger {
	if defaultLogger == nil {
		defaultLogger = NewLogger(Config{
			Level:    LevelInfo,
			UseColor: true,
		})
	}
	return defaultLogger
}

// NewLogger 创建新的日志记录器
func NewLogger(cfg Config) *Logger {
	l := &Logger{
		module:     cfg.Module,
		level:      cfg.Level,
		writer:     os.Stdout,
		useColor:   cfg.UseColor,
		jsonFormat: cfg.JSONFormat,
	}

	if cfg.LogDir != "" {
		if err := os.MkdirAll(cfg.LogDir, 0755); err == nil {
			logPath := filepath.Join(cfg.LogDir, fmt.Sprintf("app-%s.log", time.Now().Format("2006-01-02")))
			f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err == nil {
				l.fileWriter = f
			}
		}
	}

	return l
}

// WithModule 创建带有模块名的子记录器
func (l *Logger) WithModule(module string) *Logger {
	// 优先使用全局文件写入器（Init后设置），否则使用当前记录器的文件写入器
	fw := l.fileWriter
	if globalFileWriter != nil {
		fw = globalFileWriter
	}
	return &Logger{
		module:     module,
		level:      l.level,
		writer:     l.writer,
		fileWriter: fw,
		useColor:   l.useColor,
		jsonFormat: l.jsonFormat,
	}
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// log 写入日志
func (l *Logger) log(level Level, msg string, fields ...Field) {
	if level < l.level {
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05.000")
	levelName := levelNames[level]

	var output string
	if l.jsonFormat {
		entry := map[string]interface{}{
			"timestamp": now,
			"level":     levelName,
			"module":    l.module,
			"message":   msg,
		}
		for _, f := range fields {
			entry[f.Key] = f.Value
		}
		data, _ := json.Marshal(entry)
		output = string(data)
	} else {
		modulePart := ""
		if l.module != "" {
			modulePart = fmt.Sprintf(" [%s]", l.module)
		}

		fieldsPart := ""
		if len(fields) > 0 {
			for _, f := range fields {
				fieldsPart += fmt.Sprintf(" %s=%v", f.Key, f.Value)
			}
		}

		if l.useColor {
			color := levelColors[level]
			output = fmt.Sprintf("%s | %s%s%s%s | %s%s",
				now, color, levelName, colorReset, modulePart, msg, fieldsPart)
		} else {
			output = fmt.Sprintf("%s | %s%s | %s%s",
				now, levelName, modulePart, msg, fieldsPart)
		}
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	fmt.Fprintln(l.writer, output)

	activeFileWriter := l.fileWriter
	if activeFileWriter == nil {
		activeFileWriter = globalFileWriter
	}

	if activeFileWriter != nil {
		// 文件输出不带颜色
		modulePart := ""
		if l.module != "" {
			modulePart = fmt.Sprintf(" [%s]", l.module)
		}
		fieldsPart := ""
		if len(fields) > 0 {
			for _, f := range fields {
				fieldsPart += fmt.Sprintf(" %s=%v", f.Key, f.Value)
			}
		}
		fileOutput := fmt.Sprintf("%s | %s%s | %s%s\n",
			now, levelName, modulePart, msg, fieldsPart)
		fmt.Fprint(activeFileWriter, fileOutput)
	}
}

// Debug 输出调试日志
func (l *Logger) Debug(msg string, fields ...Field) {
	l.log(LevelDebug, msg, fields...)
}

// Info 输出信息日志
func (l *Logger) Info(msg string, fields ...Field) {
	l.log(LevelInfo, msg, fields...)
}

// Warn 输出警告日志
func (l *Logger) Warn(msg string, fields ...Field) {
	l.log(LevelWarn, msg, fields...)
}

// Error 输出错误日志
func (l *Logger) Error(msg string, fields ...Field) {
	l.log(LevelError, msg, fields...)
}

// ErrorErr 输出带错误对象的错误日志
func (l *Logger) ErrorErr(msg string, err error, fields ...Field) {
	fields = append(fields, Field{Key: "error", Value: err.Error()})
	l.log(LevelError, msg, fields...)
}

// Fatal 输出致命错误日志并退出
func (l *Logger) Fatal(msg string, fields ...Field) {
	l.log(LevelError, msg, fields...)
	os.Exit(1)
}

// FatalErr 输出带错误对象的致命错误日志并退出
func (l *Logger) FatalErr(msg string, err error, fields ...Field) {
	fields = append(fields, Field{Key: "error", Value: err.Error()})
	l.log(LevelError, msg, fields...)
	os.Exit(1)
}

// WithFields 创建带有预设字段的日志记录器
func (l *Logger) WithFields(fields ...Field) *FieldLogger {
	return &FieldLogger{
		logger: l,
		fields: fields,
	}
}

// FieldLogger 带字段的日志记录器
type FieldLogger struct {
	logger *Logger
	fields []Field
}

func (fl *FieldLogger) Debug(msg string) {
	fl.logger.log(LevelDebug, msg, fl.fields...)
}

func (fl *FieldLogger) Info(msg string) {
	fl.logger.log(LevelInfo, msg, fl.fields...)
}

func (fl *FieldLogger) Warn(msg string) {
	fl.logger.log(LevelWarn, msg, fl.fields...)
}

func (fl *FieldLogger) Error(msg string) {
	fl.logger.log(LevelError, msg, fl.fields...)
}

func (fl *FieldLogger) ErrorErr(msg string, err error) {
	fl.fields = append(fl.fields, Field{Key: "error", Value: err.Error()})
	fl.logger.log(LevelError, msg, fl.fields...)
}

// ========== 全局便捷函数 ==========

func Debug(msg string, fields ...Field) {
	GetLogger().Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	GetLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	GetLogger().Warn(msg, fields...)
}

func Error(msg string, fields ...Field) {
	GetLogger().Error(msg, fields...)
}

func ErrorErr(msg string, err error, fields ...Field) {
	GetLogger().ErrorErr(msg, err, fields...)
}

func Fatal(msg string, fields ...Field) {
	GetLogger().Fatal(msg, fields...)
}

func FatalErr(msg string, err error, fields ...Field) {
	GetLogger().FatalErr(msg, err, fields...)
}

func WithModule(module string) *Logger {
	return GetLogger().WithModule(module)
}

func WithFields(fields ...Field) *FieldLogger {
	return GetLogger().WithFields(fields...)
}

// Close 关闭日志记录器
func Close() {
	if defaultLogger != nil && defaultLogger.fileWriter != nil {
		defaultLogger.fileWriter.(*os.File).Close()
	}
}

// ========== 辅助函数 ==========

// F 创建日志字段
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// String 获取日志级别字符串
func (l Level) String() string {
	return levelNames[l]
}

// ParseLevel 解析日志级别字符串
func ParseLevel(s string) Level {
	switch s {
	case "debug", "DEBUG":
		return LevelDebug
	case "info", "INFO":
		return LevelInfo
	case "warn", "WARN", "warning", "WARNING":
		return LevelWarn
	case "error", "ERROR":
		return LevelError
	default:
		return LevelInfo
	}
}

// init 默认初始化
func init() {
	// 如果没有显式初始化，使用默认配置
	if defaultLogger == nil {
		defaultLogger = &Logger{
			level:    LevelInfo,
			writer:   os.Stdout,
			useColor: true,
		}
	}
}

// 替代标准库 log 的使用
var stdLog = log.New(os.Stdout, "", 0)
