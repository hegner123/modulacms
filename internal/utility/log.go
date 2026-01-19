package utility

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LogLevel defines the severity of a log message
type LogLevel int

const (
	BLANK LogLevel = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
)

func NewLogFile() *os.File {
	logField, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println(err)
	}
	return logField
}

// Logger represents a simple structured logger with levels
type Logger struct {
	level   LogLevel
	prefix  string
	logFile *os.File
}

var DefaultLogger = NewLogger(INFO)

// NewLogger creates a new logger with the specified minimum level
func NewLogger(level LogLevel) *Logger {
	f := NewLogFile()
	return &Logger{
		level:   level,
		prefix:  "",
		logFile: f,
	}
}

// WithPrefix creates a new logger with the same level but a custom prefix
func (l *Logger) WithPrefix(prefix string) *Logger {
	return &Logger{
		level:  l.level,
		prefix: prefix,
	}
}

// SetLevel changes the logger's minimum level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

var (
	// Level badges
	blankStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	debugStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#AF87FF")).Bold(true)                                       // Magenta
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#87D787")).Bold(true)                                       // Green
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAF5F")).Bold(true)                                       // Yellow
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F5F")).Bold(true)                                       // Red
	fatalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F5F")).Background(lipgloss.Color("#5F0000")).Bold(true) // Red with darker background

	// Other log elements
	timestampStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#767676"))              // Gray
	fileInfoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#5FAFD7"))              // Blue
	prefixStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#AFAFD7"))              // Light purple
	errMsgStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F5F")).Italic(true) // Red italic
)

// Helper functions that adapt the lipgloss.Style.Render method to our required signature
func styleWrapper(style lipgloss.Style) func(string) string {
	return func(s string) string {
		return style.Render(s)
	}
}

// levelStyleMap maps LogLevel to its corresponding style function
var levelStyleMap = map[LogLevel]LogLevelStyle{
	BLANK: {LevelName: "BLANK", Style: styleWrapper(blankStyle)},
	DEBUG: {LevelName: "DEBUG", Style: styleWrapper(debugStyle)},
	INFO:  {LevelName: "INFO", Style: styleWrapper(infoStyle)},
	WARN:  {LevelName: "WARN", Style: styleWrapper(warnStyle)},
	ERROR: {LevelName: "ERROR", Style: styleWrapper(errorStyle)},
	FATAL: {LevelName: "FATAL", Style: styleWrapper(fatalStyle)},
}

// formatLogMessage creates a standardized log entry with timestamp, file/line info, and message
func formatLogMessage(level LogLevel, message string, err error, args ...any) string {
	if level == BLANK {
		if args == nil {

			return fmt.Sprintln(message)
		}
		fmt.Println("Blank Level")
		return fmt.Sprintf(message+"\n", args)
	}
	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	fileInfo := "unknown:0"
	if ok {
		fileInfo = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}

	// Format timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	// Get level style
	logStyle, ok := levelStyleMap[level]
	if !ok {
		logStyle = LogLevelStyle{LevelName: "UNKNOWN", Style: styleWrapper(debugStyle)}
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
	logEntry.WriteString(fmt.Sprintf("%s ", logStyle.Style(fmt.Sprintf("[%s]", logStyle.LevelName))))
	logEntry.WriteString(fmt.Sprintf("%s ", timestampStyle.Render(timestamp)))
	logEntry.WriteString(fmt.Sprintf("%s ", fileInfoStyle.Render(fileInfo)))

	if len(DefaultLogger.prefix) > 0 {
		logEntry.WriteString(fmt.Sprintf("%s ", prefixStyle.Render(fmt.Sprintf("[%s]", DefaultLogger.prefix))))
	}

	logEntry.WriteString(fullMessage)

	if err != nil {
		logEntry.WriteString(fmt.Sprintf(": %s", errMsgStyle.Render(err.Error())))
	}

	return logEntry.String()
}

// Blank logs a raw message
func (l *Logger) Blank(message string, args ...any) {
	if l.level <= BLANK {
		fmt.Println(formatLogMessage(BLANK, message, nil, args...))
	}
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

// Fblank logs a raw message to a file
func (l *Logger) Fblank(message string, args ...any) {
	if l.level <= BLANK {
		_, err := fmt.Fprintln(l.logFile, formatLogMessage(BLANK, message, nil, args...))
		if err != nil {
			DefaultLogger.Error("", err)
		}
	}
}

// Fdebug logs a debug message to a file
func (l *Logger) Fdebug(message string, args ...any) {
	if l.level <= DEBUG {
		_, err := fmt.Fprintln(l.logFile, formatLogMessage(DEBUG, message, nil, args...))
		if err != nil {
			DefaultLogger.Error("", err)
		}
	}
}

// Finfo logs an informational message to a file
func (l *Logger) Finfo(message string, args ...any) {
	if l.level <= INFO {
		_, err := fmt.Fprintln(l.logFile, formatLogMessage(INFO, message, nil, args...))
		if err != nil {
			DefaultLogger.Error("", err)
		}
	}
}

// Fwarn logs a warning message to a file
func (l *Logger) Fwarn(message string, err error, args ...any) {
	if l.level <= WARN {
		_, writeErr := fmt.Fprintln(l.logFile, formatLogMessage(WARN, message, err, args...))
		if writeErr != nil {
			DefaultLogger.Error("Failed to write warning to log file", writeErr)
		}
	}
}

// Ferror logs an error message to a file
func (l *Logger) Ferror(message string, err error, args ...any) {
	if l.level <= ERROR {
		_, writeErr := fmt.Fprintln(l.logFile, formatLogMessage(ERROR, message, err, args...))
		if writeErr != nil {
			DefaultLogger.Error("Failed to write error to log file", writeErr)
		}
	}
}

// Ffatal logs an error message to a file and exits the program
func (l *Logger) Ffatal(message string, err error, args ...any) {
	if l.level <= FATAL {
		_, writeErr := fmt.Fprintln(l.logFile, formatLogMessage(FATAL, message, err, args...))
		if writeErr != nil {
			DefaultLogger.Error("Failed to write fatal error to log file", writeErr)
		}
		os.Exit(1)
	}
}
