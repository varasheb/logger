package logger

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"runtime"

	_ "github.com/lib/pq"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	db        *sql.DB
	logger    *log.Logger
	processID int
	createdBy string
}

func InitLogger(dbURL string, logFilePath string, createdBy string) (*Logger, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %v", err)
	}

	if pingErr := db.Ping(); pingErr != nil {
		return nil, fmt.Errorf("database ping failed: %v", pingErr)
	}

	logFile := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    10, // MB
		MaxBackups: 5,
		MaxAge:     30, // Days
		Compress:   true,
	}

	loggerInstance := log.New(logFile, "", log.LstdFlags|log.Lshortfile)

	pid := os.Getpid()

	return &Logger{db: db, logger: loggerInstance, processID: pid, createdBy: createdBy}, nil
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
	query := `INSERT INTO fotadevicelogs (processid, deviceid, fileid, loglevel, status, createdby, error_details, createdat)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, EXTRACT(EPOCH FROM NOW()) * 1000)`

	_, dbErr := l.db.Exec(query, l.processID, deviceID, fileID, logLevel, status, l.createdBy, errorDetails)
	if dbErr != nil {
		l.logger.Println("Failed to insert log into DB:", dbErr)
	}
}

func (l *Logger) Log(deviceID, fileID, logLevel, status string, err error) {
	logMessage := fmt.Sprintf("[%s] Device: %s, File: %s, Status: %s", logLevel, deviceID, fileID, status)
	l.logger.Println(logMessage)

	if err != nil {
		stackTrace := captureStackTrace(err)
		l.logger.Println("Error details:", stackTrace)
	}
	l.LogToDB(deviceID, fileID, logLevel, status, err)
}
