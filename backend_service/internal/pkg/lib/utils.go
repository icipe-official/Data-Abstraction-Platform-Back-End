package lib

import (
	"log"
	"os"
	"strconv"
)

func Log(logLevel int, logSection, logMessage string) {
	if logLevel < LOG_LEVEL {
		return
	}
	logLevelType := "DEBUG"
	if logLevel == LOG_INFO {
		logLevelType = "INFO"
	} else if logLevel == LOG_WARNING {
		logLevelType = "WARNING"
	} else if logLevel == LOG_ERROR {
		logLevelType = "ERROR"
	} else if logLevel == LOG_FATAL {
		logLevelType = "FATAL"
	}
	log.Printf("%v | %v | %v\n", logLevelType, logSection, logMessage)
	if logLevel == LOG_FATAL {
		os.Exit(1)
	}
}

func getLogLevel() int {
	if lv, err := strconv.Atoi(os.Getenv("LOG_LEVEL")); err != nil {
		return 1
	} else {
		return lv
	}
}
