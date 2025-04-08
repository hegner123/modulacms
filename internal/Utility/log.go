package utility

import (
	"embed"
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
	DEBUG LogLevel = iota
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

var DefaultLogger = NewLogger(DEBUG)

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

var (
	// Level badges
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
	DEBUG: {LevelName: "DEBUG", Style: styleWrapper(debugStyle)},
	INFO:  {LevelName: "INFO", Style: styleWrapper(infoStyle)},
	WARN:  {LevelName: "WARN", Style: styleWrapper(warnStyle)},
	ERROR: {LevelName: "ERROR", Style: styleWrapper(errorStyle)},
	FATAL: {LevelName: "FATAL", Style: styleWrapper(fatalStyle)},
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
func (l *Logger) Fdebug(message string, args ...any) {
	if l.level <= DEBUG {
		_, err := fmt.Fprintln(l.logFile, formatLogMessage(DEBUG, message, nil, args...))
		if err != nil {
			DefaultLogger.Error("", err)
		}
	}
}

// Info logs an informational message
func (l *Logger) Finfo(message string, args ...any) {
	if l.level <= INFO {
		_, err := fmt.Fprintln(l.logFile, formatLogMessage(INFO, message, nil, args...))
		if err != nil {
			DefaultLogger.Error("", err)
		}
	}
}

// Warn logs a warning message
func (l *Logger) Fwarn(message string, err error, args ...any) {
	if l.level <= WARN {
		_, err := fmt.Fprintln(l.logFile, formatLogMessage(WARN, message, err, args...))
		if err != nil {
			DefaultLogger.Error("", err)
		}
	}
}

// Error logs an error message
func (l *Logger) Ferror(message string, err error, args ...any) {
	if l.level <= ERROR {
		_, err := fmt.Fprintln(l.logFile, formatLogMessage(ERROR, message, err, args...))
		if err != nil {
			DefaultLogger.Error("", err)
		}
	}
}

// Fatal logs an error message and exits the program
func (l *Logger) Ffatal(message string, err error, args ...any) {
	if l.level <= FATAL {
		_, err := fmt.Fprintln(l.logFile, formatLogMessage(FATAL, message, err, args...))
		if err != nil {
			DefaultLogger.Error("", err)
		}
		os.Exit(1)
	}
}
