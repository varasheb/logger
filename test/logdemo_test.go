package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/varasheb/logger"
)

func testLogger() {
	dbURL := "postgresql://logger_owner:1234567890@localhost:5432/logger"
	logFilePath := "app.log"
	createdBy := "TestUser"

	loggerI, err := logger.InitLogger(dbURL, "rivertest", createdBy, logFilePath, "fotalogs")
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer loggerI.Close()

	// Log different severity levels
	loggerI.Log("0101FCED8C5C2AE2", "00064467DD629E36C838C3E94F20A490A96B4DB084FA300C3DD64D5D45949DE9", "INFO", "System is running smoothly", map[string]interface{}{"testmeta": "testvalue"}, nil)
	time.Sleep(2 * time.Minute)
	loggerI.Log("0102474A083CA0EE", "00064467DD629E36C838C3E94F20A490A96B4DB084FA300C3DD64D3D45949DE9", "WARN", "Memory usage is high", nil, fmt.Errorf("memory usage at 90%"))
	// time.Sleep(2 * time.Minute)
	loggerI.Log("0101FCED8C5C2AE2", "00064467DD629E36C838C3E94F20A490A96B4DB084FA300C3DD64D2D45949DE9", "ERROR", "Database connection failed", nil, fmt.Errorf("connection timeout"))
	loggerI.Log("0102474A083CA0EE", "00064467DD629E36C838C3E94F20A490A96B4DB084FA300C3DD64D2D45949DE9", "DEBUG", "System crash detected", map[string]interface{}{"testmeta": "testvalue"}, fmt.Errorf("kernel panic"))

	fmt.Println("Logger test completed. Check 'app.log' and database for logs.")
	// Close the logger
	loggerI.Close()
}

func main() {
	fmt.Println("Process ID:", os.Getpid())
	testLogger()
}
