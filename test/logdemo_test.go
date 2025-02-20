package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/varasheb/logger"
)

func testLogger() {
	dbURL := "postgresql://testing_owner:rilHO3obSc7X@ep-weathered-credit-a1p4w5k1-pooler.ap-southeast-1.aws.neon.tech/Dmt?sslmode=require"
	logFilePath := "app.log"
	createdBy := "TestUser"

	// Initialize logger
	loggerI, err := logger.InitLogger(dbURL, logFilePath, createdBy, "fotalogs")
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer loggerI.Close()

	// Log different severity levels

	loggerI.Log("device001", "fileABC", "INFO", "System is running smoothly", nil)
	time.Sleep(2 * time.Minute)
	loggerI.Log("device002", "fileDEF", "WARNING", "Memory usage is high", fmt.Errorf("memory usage at 90%"))
	// time.Sleep(2 * time.Minute)
	loggerI.Log("device003", "fileXYZ", "ERROR", "Database connection failed", fmt.Errorf("connection timeout"))
	loggerI.Log("device004", "filePQR", "CRITICAL", "System crash detected", fmt.Errorf("kernel panic"))

	fmt.Println("Logger test completed. Check 'app.log' and database for logs.")
}

func main() {
	fmt.Println("Process ID:", os.Getpid())
	testLogger()
	time.Sleep(1 * time.Second)

}
