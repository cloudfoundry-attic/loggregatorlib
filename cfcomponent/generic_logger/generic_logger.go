package generic_logger

import (
	"log"
)

type GenericLogger interface {
	Fatalf(string, ...interface{})
	Errorf(string, ...interface{})
	Debugf(string, ...interface{})

}

type defaultGenericLogger struct{}

func NewDefaultGenericLogger() GenericLogger {
	return defaultGenericLogger{}
}

func (defaultGenericLogger) Fatalf(format string, args... interface{}) {
	log.Fatalf(format, args...)
}

func (defaultGenericLogger) Errorf(format string, args... interface{}) {
	log.Printf("ERROR: " + format, args...)
}

func (defaultGenericLogger) Debugf(format string, args... interface{}) {
	log.Printf("DEBUG: " + format, args...)
}
