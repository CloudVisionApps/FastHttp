package utils

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	webServerLogger *log.Logger
	adminLogger     *log.Logger
	errorLogger     *log.Logger
	loggerMutex     sync.Mutex
)

// InitLoggers initializes loggers based on config
func InitLoggers(logFile, adminLogFile, errorLogFile string) error {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	// Web server logger
	var webWriter io.Writer = os.Stdout
	if logFile != "" {
		// Ensure directory exists
		dir := filepath.Dir(logFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		webWriter = file
	}
	webServerLogger = log.New(webWriter, "", log.LstdFlags)

	// Admin API logger
	var adminWriter io.Writer = os.Stdout
	if adminLogFile != "" {
		dir := filepath.Dir(adminLogFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		file, err := os.OpenFile(adminLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		adminWriter = file
	}
	adminLogger = log.New(adminWriter, "", log.LstdFlags)

	// Error logger
	var errorWriter io.Writer = os.Stderr
	if errorLogFile != "" {
		dir := filepath.Dir(errorLogFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		file, err := os.OpenFile(errorLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		errorWriter = file
	}
	errorLogger = log.New(errorWriter, "", log.LstdFlags)

	return nil
}

// WebServerLog logs a message for the web server
func WebServerLog(format string, v ...interface{}) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	if webServerLogger != nil {
		webServerLogger.Printf(format, v...)
	} else {
		log.Printf(format, v...)
	}
}

// AdminLog logs a message for the admin API
func AdminLog(format string, v ...interface{}) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	if adminLogger != nil {
		adminLogger.Printf(format, v...)
	} else {
		log.Printf(format, v...)
	}
}

// ErrorLog logs an error message
func ErrorLog(format string, v ...interface{}) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	if errorLogger != nil {
		errorLogger.Printf(format, v...)
	} else {
		log.Printf(format, v...)
	}
}

// GetWebServerLogger returns the web server logger
func GetWebServerLogger() *log.Logger {
	return webServerLogger
}

// GetAdminLogger returns the admin logger
func GetAdminLogger() *log.Logger {
	return adminLogger
}

// GetErrorLogger returns the error logger
func GetErrorLogger() *log.Logger {
	return errorLogger
}
