package logger

import (
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

// Initialize sets up the logger to write to a file
func Initialize() error {
	// Create logs directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	logDir := filepath.Join(homeDir, ".hacktober", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logDir, "hacktober-"+timestamp+".log")

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	// Configure zerolog
	zerolog.TimeFieldFormat = time.RFC3339
	Logger = zerolog.New(file).
		Level(zerolog.DebugLevel).
		With().
		Timestamp().
		Caller().
		Logger()

	// Also set global logger
	log.Logger = Logger

	Logger.Info().
		Str("version", "1.0.0").
		Str("log_file", logFile).
		Msg("Hacktoberfest CLI started")

	return nil
}

// Debug logs a debug message
func Debug(msg string) {
	Logger.Debug().Msg(msg)
}

// Info logs an info message
func Info(msg string) {
	Logger.Info().Msg(msg)
}

// Warn logs a warning message
func Warn(msg string) {
	Logger.Warn().Msg(msg)
}

// Error logs an error message
func Error(msg string) {
	Logger.Error().Msg(msg)
}

// ErrorWithErr logs an error with error details
func ErrorWithErr(msg string, err error) {
	Logger.Error().Err(err).Msg(msg)
}

// WithFields creates a logger with additional fields
func WithFields(fields map[string]interface{}) zerolog.Logger {
	event := Logger.With()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	return event.Logger()
}

// LogAPIRequest logs GitHub API requests
func LogAPIRequest(endpoint string, query string, statusCode int, duration time.Duration) {
	Logger.Info().
		Str("endpoint", endpoint).
		Str("query", query).
		Int("status_code", statusCode).
		Dur("duration", duration).
		Msg("GitHub API request")
}

// LogRepoSearch logs repository search results
func LogRepoSearch(query string, totalFound int, returned int, languages []string) {
	Logger.Info().
		Str("query", query).
		Int("total_found", totalFound).
		Int("returned", returned).
		Strs("languages", languages).
		Msg("Repository search completed")
}

// LogIssueSearch logs issue search results
func LogIssueSearch(repo string, totalFound int, returned int) {
	Logger.Info().
		Str("repository", repo).
		Int("total_found", totalFound).
		Int("returned", returned).
		Msg("Issue search completed")
}

// GetLogLocation returns the current log file path
func GetLogLocation() string {
	homeDir, _ := os.UserHomeDir()
	timestamp := time.Now().Format("2006-01-02")
	return filepath.Join(homeDir, ".hacktober", "logs", "hacktober-"+timestamp+".log")
}
