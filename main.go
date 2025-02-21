package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	db          *pgxpool.Pool
	logger      *log.Logger
	processID   string
	createdBy   string
	schema      string
	processName string
}

/*
InitLogger initializes the logger with a database connection, file logging, and a process ID.

Parameters:
- `dbURL`: Database connection string.
- `processName`: The name of the process running the logger.
- `createdBy`: Identifier for the user or service creating logs.
- `logFilePath`: Path to the log file.
- `schema`: (Optional) Database schema name; defaults to `"public"`.

Returns:
- A pointer to a `Logger` instance.
- An error if initialization fails.
*/
func InitLogger(dbURL, processName, createdBy, logFilePath, schema string) (*Logger, error) {
	// Ensure the schema is valid, defaulting to "public"
	var defaultSchema string
	if len(schema) <= 3 {
		defaultSchema = "public"
	} else {
		defaultSchema = schema
	}

	// Default log file path if not provided
	if logFilePath == "" {
		execPath, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %v", err)
		}
		logFilePath = filepath.Join(execPath, "applogs.log")
	} else {
		if !strings.HasSuffix(logFilePath, ".log") {
			logFilePath = filepath.Join(logFilePath, "applogs.log")
		}
	}

	// Initialize database connection
	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if pingErr := db.Ping(ctx); pingErr != nil {
		return nil, fmt.Errorf("database ping failed: %v", pingErr)
	}

	// Create schema if it doesn't exist
	createSchemaQuery := fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s;`, defaultSchema)
	_, err = db.Exec(ctx, createSchemaQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema: %v", err)
	}

	// Create log table if it does not exist
	createTableQuery := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s.fotadevicelogs (
		processid TEXT NOT NULL,
		processname TEXT NOT NULL,
		deviceid TEXT NOT NULL,
		fileid TEXT NOT NULL,
		loglevel TEXT NOT NULL,
		status TEXT NOT NULL,
		errormessage TEXT, 
		metadata JSONB DEFAULT '{}',
		createdby TEXT NOT NULL, 
		createdat BIGINT DEFAULT CAST(extract(epoch FROM NOW()) * 1000 AS BIGINT) NOT NULL,
		PRIMARY KEY (processid, deviceid, fileid , processname, createdat));`, defaultSchema)

	_, err = db.Exec(ctx, createTableQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to create log table: %v", err)
	}

	// Configure file-based logging
	logFile := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    10,   // MB
		MaxBackups: 5,    // Number of old logs to retain
		MaxAge:     30,   // Days
		Compress:   true, // Enable compression
	}

	loggerInstance := log.New(logFile, "", log.LstdFlags|log.Lshortfile)
	pid := os.Getpid()

	return &Logger{
		db:          db,
		logger:      loggerInstance,
		processID:   fmt.Sprintf("%d", pid),
		processName: processName,
		createdBy:   createdBy,
		schema:      defaultSchema,
	}, nil
}

/*
captureStackTrace captures the stack trace of an error.

Returns:
- A formatted string containing the error message and its source file/line number.
*/
func captureStackTrace(err error) string {
	if err == nil {
		return ""
	}

	_, file, line, ok := runtime.Caller(2)
	if ok {
		return fmt.Sprintf("Error: %v | File: %s | Line: %d", err, file, line)
	}
	return fmt.Sprintf("Error: %v", err)
}

/*
LogToDB inserts a log entry into the database.
*/
func (l *Logger) LogToDB(deviceID, fileID, logLevel, status string, metadata interface{}, err error) {
	if l.db == nil {
		l.logger.Println("DB logging skipped: No database connection")
		return
	}

	// Normalize log level and status to lowercase
	logLevel = strings.ToLower(logLevel)
	status = strings.ToLower(status)

	// Validate input parameters
	if len(deviceID) != 16 || len(fileID) != 64 {
		l.logger.Println("Invalid log: Device ID must be 16 digits, File ID must be 64 digits")
		return
	}

	validLogLevels := map[string]bool{"info": true, "warn": true, "error": true, "debug": true}
	if !validLogLevels[logLevel] {
		l.logger.Println("Invalid log: Log level must be one of: info, warn, error, debug")
		return
	}

	errorDetails := captureStackTrace(err)

	compressedMetadata, err := json.Marshal(metadata)
	if err != nil {
		l.logger.Printf("Failed to compress metadata: %v\n", err)
		return
	}

	query := fmt.Sprintf(`
		INSERT INTO %s.fotadevicelogs (processid, processname, deviceid, fileid, loglevel, status, createdby, errormessage, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`, l.schema)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, dbErr := l.db.Exec(ctx, query, l.processID, l.processName, deviceID, fileID, logLevel, status, l.createdBy, errorDetails, compressedMetadata)
	if dbErr != nil {
		l.logger.Printf("Failed to insert log into DB: %v\n", dbErr)
	}
}

/*
Close closes the database connection.
*/
func (l *Logger) Close() {
	if l.db != nil {
		l.db.Close()
	}
}
