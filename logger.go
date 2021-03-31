package main

import "fmt"

// LogLevel indicates log level
type LogLevel int

const (
	info LogLevel = iota
	warn
	err
)

// Info ...
func Info(message string) {
	Log(info, message)
}

// Warn ...
func Warn(message string) {
	Log(warn, message)
}

// Error ...
func Error(message string) {
	Log(err, message)
}

// Infof ...
func Infof(format string, a ...interface{}) {
	message := fmt.Sprintf(format, a...)
	Log(info, message)
}

// Warnf ...
func Warnf(format string, a ...interface{}) {
	message := fmt.Sprintf(format, a...)
	Log(warn, message)
}

// Errorf ...
func Errorf(format string, a ...interface{}) {
	message := fmt.Sprintf(format, a...)
	Log(err, message)
}

// Log ...
func Log(logLevel LogLevel, message string) {
	fmt.Println(message)
}
