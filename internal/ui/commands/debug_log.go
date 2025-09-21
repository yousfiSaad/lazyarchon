package commands

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

var (
	debugLogger *log.Logger
	logFile     *os.File
	logOnce     sync.Once
)

// InitDebugLog initializes debug logging to file
func InitDebugLog() {
	logOnce.Do(func() {
		var err error
		logFile, err = os.OpenFile("/tmp/lazyarchon-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return // Fail silently if can't create log file
		}

		debugLogger = log.New(logFile, "", 0)
		debugLogger.Println("=== LazyArchon Debug Session Started ===")
		debugLogger.Println("Time:", time.Now().Format("2006-01-02 15:04:05"))
	})
}

// DebugLog writes a debug message to the log file
func DebugLog(format string, args ...interface{}) {
	InitDebugLog()
	if debugLogger != nil {
		message := fmt.Sprintf(format, args...)
		debugLogger.Printf("[DEBUG] %s", message)
	}
}

// CloseDebugLog closes the debug log file
func CloseDebugLog() {
	if logFile != nil {
		debugLogger.Println("=== Debug Session Ended ===")
		logFile.Close()
	}
}