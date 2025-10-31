package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.etcd.io/bbolt"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel parses a string to LogLevel
func ParseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

// Logger represents a logger instance
type Logger struct {
	level    LogLevel
	output   io.Writer
	file     *os.File
	db       *bbolt.DB
	config   LogConfig
	features map[string]bool
}

// LogConfig represents logging configuration
type LogConfig struct {
	Enabled  bool
	MaxDays  int
	MaxSize  int // in MB
	Compress bool
	LogDir   string
	Features []string
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Feature   string                 `json:"feature"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// NewLogger creates a new logger instance
func NewLogger(config LogConfig) (*Logger, error) {
	logger := &Logger{
		level:    INFO,
		config:   config,
		features: make(map[string]bool),
	}

	// Set up feature filtering
	for _, feature := range config.Features {
		logger.features[feature] = true
	}

	// Create log directory if it doesn't exist
	if err := os.MkdirAll(config.LogDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log database
	dbPath := filepath.Join(config.LogDir, "logs.db")
	db, err := bbolt.Open(dbPath, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("failed to open log database: %w", err)
	}

	logger.db = db

	// Initialize buckets
	if err := logger.initBuckets(); err != nil {
		return nil, fmt.Errorf("failed to initialize log buckets: %w", err)
	}

	// Set up log rotation
	if config.Enabled {
		go logger.startRotation()
	}

	return logger, nil
}

// initBuckets initializes the log database buckets
func (l *Logger) initBuckets() error {
	return l.db.Update(func(tx *bbolt.Tx) error {
		// Create main logs bucket
		if _, err := tx.CreateBucketIfNotExists([]byte("logs")); err != nil {
			return err
		}

		// Create feature-specific buckets
		for feature := range l.features {
			if _, err := tx.CreateBucketIfNotExists([]byte("feature_" + feature)); err != nil {
				return err
			}
		}

		return nil
	})
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// SetOutput sets the output writer
func (l *Logger) SetOutput(w io.Writer) {
	l.output = w
}

// Debug logs a debug message
func (l *Logger) Debug(feature, message string, data ...map[string]interface{}) {
	l.log(DEBUG, feature, message, data...)
}

// Info logs an info message
func (l *Logger) Info(feature, message string, data ...map[string]interface{}) {
	l.log(INFO, feature, message, data...)
}

// Warn logs a warning message
func (l *Logger) Warn(feature, message string, data ...map[string]interface{}) {
	l.log(WARN, feature, message, data...)
}

// Error logs an error message
func (l *Logger) Error(feature, message string, data ...map[string]interface{}) {
	l.log(ERROR, feature, message, data...)
}

// log logs a message with the specified level
func (l *Logger) log(level LogLevel, feature, message string, data ...map[string]interface{}) {
	if level < l.level {
		return
	}

	// Check feature filtering
	if len(l.features) > 0 && !l.features[feature] {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Feature:   feature,
		Message:   message,
	}

	if len(data) > 0 {
		entry.Data = data[0]
	}

	// Write to console if output is set
	if l.output != nil {
		l.writeToConsole(entry)
	}

	// Write to database
	if l.config.Enabled {
		l.writeToDatabase(entry)
	}
}

// writeToConsole writes a log entry to console
func (l *Logger) writeToConsole(entry LogEntry) {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
	level := entry.Level.String()
	feature := entry.Feature

	// Color coding
	var color string
	switch entry.Level {
	case DEBUG:
		color = "\033[36m" // Cyan
	case INFO:
		color = "\033[32m" // Green
	case WARN:
		color = "\033[33m" // Yellow
	case ERROR:
		color = "\033[31m" // Red
	}

	reset := "\033[0m"

	message := fmt.Sprintf("%s[%s] %s %s: %s%s\n",
		color, timestamp, level, feature, entry.Message, reset)

	if l.output != nil {
		l.output.Write([]byte(message))
	} else {
		fmt.Print(message)
	}
}

// writeToDatabase writes a log entry to the database
func (l *Logger) writeToDatabase(entry LogEntry) {
	l.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("logs"))
		if bucket == nil {
			return fmt.Errorf("logs bucket not found")
		}

		// Create key with timestamp and random suffix for uniqueness
		key := fmt.Sprintf("%d_%d", entry.Timestamp.UnixNano(), time.Now().UnixNano())

		// Serialize entry
		data, err := entry.serialize()
		if err != nil {
			return err
		}

		return bucket.Put([]byte(key), data)
	})
}

// serialize serializes a log entry to JSON
func (e *LogEntry) serialize() ([]byte, error) {
	// Simple JSON serialization (in production, use proper JSON library)
	return []byte(fmt.Sprintf(`{"timestamp":"%s","level":"%s","feature":"%s","message":"%s"}`,
		e.Timestamp.Format(time.RFC3339), e.Level.String(), e.Feature, e.Message)), nil
}

// startRotation starts the log rotation process
func (l *Logger) startRotation() {
	ticker := time.NewTicker(24 * time.Hour) // Check daily
	defer ticker.Stop()

	for range ticker.C {
		l.rotateLogs()
	}
}

// rotateLogs rotates old log entries
func (l *Logger) rotateLogs() {
	cutoff := time.Now().AddDate(0, 0, -l.config.MaxDays)

	l.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("logs"))
		if bucket == nil {
			return nil
		}

		cursor := bucket.Cursor()
		for key, _ := cursor.First(); key != nil; key, _ = cursor.Next() {
			// Parse timestamp from key (simplified)
			// In production, store proper timestamps
			if time.Now().After(cutoff) {
				break
			}

			cursor.Delete()
		}

		return nil
	})
}

// GetLogs retrieves logs for a specific feature
func (l *Logger) GetLogs(feature string, limit int) ([]LogEntry, error) {
	var entries []LogEntry

	err := l.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("logs"))
		if bucket == nil {
			return fmt.Errorf("logs bucket not found")
		}

		cursor := bucket.Cursor()
		count := 0

		for key, value := cursor.Last(); key != nil && count < limit; key, value = cursor.Prev() {
			// Parse entry (simplified)
			entry := LogEntry{}
			// In production, use proper JSON unmarshaling
			entry.Message = string(value)
			entries = append(entries, entry)
			count++
		}

		return nil
	})

	return entries, err
}

// Close closes the logger and database
func (l *Logger) Close() error {
	if l.file != nil {
		l.file.Close()
	}
	if l.db != nil {
		return l.db.Close()
	}
	return nil
}

// ClearLogs clears all logs
func (l *Logger) ClearLogs() error {
	return l.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("logs"))
		if bucket == nil {
			return nil
		}

		return bucket.ForEach(func(key, _ []byte) error {
			return bucket.Delete(key)
		})
	})
}
