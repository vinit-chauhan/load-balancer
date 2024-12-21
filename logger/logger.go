package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

var logger *log.Logger
var logLevel = LevelInfo

const (
	LevelDebug = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelPanic
)

const defaultDir = "logs"
const logFileName = "proxy.log"

func Init() {
	var logDir string

	if logDir = os.Getenv("LOG_DIR"); logDir == "" {
		logDir = defaultDir
	}

	// Create log dir.
	if _, err := os.ReadDir(logDir); err != nil {
		if err := os.Mkdir(logDir, 0755); err != nil {
			panic(fmt.Sprintf("Error initializing log dir (%s): %s", logDir, err.Error()))
		}
	}

	fd, err := os.OpenFile(logDir+"/"+logFileName, os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		panic(fmt.Sprintf("Error creating log file (%s): %s", logDir+"/"+logFileName, err.Error()))
	}

	logger = log.New(fd, "[proxy]", log.LstdFlags|log.LUTC)
}

func SetLogLevel(level int) {
	logLevel = level
}

func addLog(level int, msg string) {
	if level >= logLevel {
		logger.Println(msg)
	}
}

func Debug(tag string, msg string, keys ...fmt.Stringer) {
	typ := "Debug"

	if len(keys) == 0 {
		addLog(LevelDebug, fmt.Sprintf("[%s] [%s] %s", typ, tag, msg))
		return
	}

	sb := strings.Builder{}
	for _, key := range keys {
		sb.WriteString(key.String() + ", ")
	}

	addLog(LevelDebug, fmt.Sprintf("[%s] [%s] {%s} %s", typ, tag, sb.String(), msg))
}

func Info(tag string, msg string) {
	addLog(LevelInfo, fmt.Sprintf("[%s] [%s] %s", "Info", tag, msg))
}

func Warn(tag string, msg string) {
	addLog(LevelWarn, fmt.Sprintf("[%s] [%s] %s", "Warn", tag, msg))
}

func Error(tag string, msg string) {
	addLog(LevelError, fmt.Sprintf("[%s] [%s] %s", "Error", tag, msg))
}

func Panic(tag string, msg string, keys ...fmt.Stringer) {
	if len(keys) == 0 {
		addLog(LevelDebug, fmt.Sprintf("[%s] [%s] %s", "Panic", tag, msg))
		return
	}

	sb := strings.Builder{}
	for _, key := range keys {
		sb.WriteString(key.String() + ", ")
	}
	addLog(LevelPanic, fmt.Sprintf("[%s] [%s] %s: %s", "Panic", tag, msg, sb.String()))
}
