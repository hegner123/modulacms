package utility

import (
	"embed"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// LogLevel defines the severity of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Logger represents a simple structured logger with levels
type Logger struct {
	level  LogLevel
	prefix string
}

var DefaultLogger = NewLogger(DEBUG)

// NewLogger creates a new logger with the specified minimum level
func NewLogger(level LogLevel) *Logger {
	return &Logger{
		level:  level,
		prefix: "",
	}
}

// WithPrefix creates a new logger with the same level but a custom prefix
func (l *Logger) WithPrefix(prefix string) *Logger {
	return &Logger{
		level:  l.level,
		prefix: prefix,
	}
}

//go:embed version.json
var vJson embed.FS

func GetVersion() (*string, error) {
	file, err := vJson.ReadFile("version.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read version file: %w", err)
	}
	versionString := string(file)
	return &versionString, nil
}

func PopError(err error) string {
	unwrappedErr := strings.Split(err.Error(), " ")
	msg := fmt.Sprint(unwrappedErr[len(unwrappedErr)-1])
	return msg
}

// LogError logs an error with context. This is the legacy method maintained for compatibility.
func LogError(message string, err error, args ...any) {
	DefaultLogger.Error(message, err, args...)
}

// formatLogMessage creates a standardized log entry with timestamp, file/line info, and message
func formatLogMessage(level LogLevel, message string, err error, args ...any) string {
	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	fileInfo := "unknown:0"
	if ok {
		fileInfo = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}

	// Format timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	// Prepare level indicator
	levelStr := "UNKNOWN"
	var levelColor ANSIForegroundColor
	switch level {
	case DEBUG:
		levelStr = "DEBUG"
		levelColor = MAGENTAF
	case INFO:
		levelStr = "INFO"
		levelColor = GREENF
	case WARN:
		levelStr = "WARN"
		levelColor = YELLOWF
	case ERROR:
		levelStr = "ERROR"
		levelColor = REDF
	case FATAL:
		levelStr = "FATAL"
		levelColor = REDF
	}

	// Format message parts
	var messageParts []string
	messageParts = append(messageParts, message)
	for _, arg := range args {
		switch v := arg.(type) {
		case fmt.Stringer:
			messageParts = append(messageParts, v.String())
		default:
			messageParts = append(messageParts, fmt.Sprintf("%+v", arg))
		}
	}
	fullMessage := strings.Join(messageParts, " ")

	// Build the log entry
	var logEntry strings.Builder
	logEntry.WriteString(fmt.Sprintf("%s[%s]%s ", levelColor, levelStr, RESET))
	logEntry.WriteString(fmt.Sprintf("%s ", timestamp))
	logEntry.WriteString(fmt.Sprintf("%s ", fileInfo))

	if len(DefaultLogger.prefix) > 0 {
		logEntry.WriteString(fmt.Sprintf("[%s] ", DefaultLogger.prefix))
	}

	logEntry.WriteString(fullMessage)

	if err != nil {
		logEntry.WriteString(fmt.Sprintf(": %s%v%s", REDF, err, RESET))
	}

	return logEntry.String()
}

// Debug logs a debug message
func (l *Logger) Debug(message string, args ...any) {
	if l.level <= DEBUG {
		fmt.Println(formatLogMessage(DEBUG, message, nil, args...))
	}
}

// Info logs an informational message
func (l *Logger) Info(message string, args ...any) {
	if l.level <= INFO {
		fmt.Println(formatLogMessage(INFO, message, nil, args...))
	}
}

// Warn logs a warning message
func (l *Logger) Warn(message string, err error, args ...any) {
	if l.level <= WARN {
		fmt.Println(formatLogMessage(WARN, message, err, args...))
	}
}

// Error logs an error message
func (l *Logger) Error(message string, err error, args ...any) {
	if l.level <= ERROR {
		fmt.Println(formatLogMessage(ERROR, message, err, args...))
	}
}

// Fatal logs an error message and exits the program
func (l *Logger) Fatal(message string, err error, args ...any) {
	if l.level <= FATAL {
		fmt.Println(formatLogMessage(FATAL, message, err, args...))
		os.Exit(1)
	}
}

// Debug logs a debug message
func (l *Logger) Fdebug(w io.Writer, message string, args ...any) {
	if l.level <= DEBUG {
		fmt.Fprintln(w, formatLogMessage(DEBUG, message, nil, args...))
	}
}

// Info logs an informational message
func (l *Logger) Finfo(w io.Writer, message string, args ...any) {
	if l.level <= INFO {
		fmt.Fprintln(w, formatLogMessage(INFO, message, nil, args...))
	}
}

// Warn logs a warning message
func (l *Logger) Fwarn(w io.Writer, message string, err error, args ...any) {
	if l.level <= WARN {
		fmt.Fprintln(w, formatLogMessage(WARN, message, err, args...))
	}
}

// Error logs an error message
func (l *Logger) Ferror(w io.Writer, message string, err error, args ...any) {
	if l.level <= ERROR {
		fmt.Fprintln(w, formatLogMessage(ERROR, message, err, args...))
	}
}

// Fatal logs an error message and exits the program
func (l *Logger) Ffatal(w io.Writer, message string, err error, args ...any) {
	if l.level <= FATAL {
		fmt.Fprintln(w, formatLogMessage(FATAL, message, err, args...))
		os.Exit(1)
	}
}

// LogHeader prints arguments in bright blue. Legacy, consider using Logger.Info with prefix.
func LogHeader(args ...any) {
	header := fmt.Sprint(args...)
	fmt.Printf("%s%s%s\n", BRIGHTBLUEF, header, RESET)
}

// LogBody prints arguments in blue. Legacy, consider using Logger.Info instead.
func LogBody(args ...any) {
	DefaultLogger.Info(fmt.Sprint(args...))
}
