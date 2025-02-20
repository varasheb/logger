package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	db        *pgxpool.Pool
	logger    *log.Logger
	processID string
	createdBy string
	schema    string
}

func InitLogger(dbURL, logFilePath, createdBy string, schema ...string) (*Logger, error) {
	defaultSchema := "public"
	if len(schema) > 0 && schema[0] != "" {
		defaultSchema = schema[0]
	}

	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if pingErr := db.Ping(ctx); pingErr != nil {
		return nil, fmt.Errorf("database ping failed: %v", pingErr)
	}

	// Configure rotating file logger
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
		db:        db,
		logger:    loggerInstance,
		processID: fmt.Sprintf("%d", pid),
		createdBy: createdBy,
		schema:    defaultSchema,
	}, nil
}

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

func (l *Logger) LogToDB(deviceID, fileID, logLevel, status string, err error) {
	if l.db == nil {
		l.logger.Println("DB logging skipped: No database connection")
		return
	}

	errorDetails := captureStackTrace(err)

	query := fmt.Sprintf(`
		INSERT INTO %s.fotadevicelogs (processid, deviceid, fileid, loglevel, status, createdby, errormessage)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`, l.schema)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, dbErr := l.db.Exec(ctx, query, l.processID, deviceID, fileID, logLevel, status, l.createdBy, errorDetails)
	if dbErr != nil {
		l.logger.Printf("Failed to insert log into DB: %v\n", dbErr)
	}
}

func (l *Logger) Log(deviceID, fileID, logLevel, status string, err error) {
	logMessage := fmt.Sprintf("[%s] Device: %s, File: %s, Status: %s", logLevel, deviceID, fileID, status)

	if err != nil {
		stackTrace := captureStackTrace(err)
		logMessage += fmt.Sprintf(" | Error details: %s", stackTrace)
	}
	l.logger.Println(logMessage)

	l.LogToDB(deviceID, fileID, logLevel, status, err)
}

func (l *Logger) Close() {
	if l.db != nil {
		l.db.Close()
	}
}
