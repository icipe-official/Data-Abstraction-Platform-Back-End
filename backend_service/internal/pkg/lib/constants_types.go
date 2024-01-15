package lib

const OPTS_SPLIT = "%!!%"

const (
	LOG_DEBUG   int = 0
	LOG_INFO    int = 1
	LOG_WARNING int = 2
	LOG_ERROR   int = 3
	LOG_FATAL   int = 4
)

var LOG_LEVEL = getLogLevel()
