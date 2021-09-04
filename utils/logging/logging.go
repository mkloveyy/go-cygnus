// logger configuration and helpers, currently share same file logger

package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"go.elastic.co/apm/module/apmlogrus"
)

var (
	LogDir = "./logs/" // conventionally injected by build -X

	LogFileNames = make(map[string]*os.File)
	loggerCache  = make(map[string]*ConvenientErrorLogger)
)

func Finalize() {
	for _, file := range LogFileNames {
		_ = file.Close()
	}
}

func init() {
	if err := os.MkdirAll(LogDir, os.ModePerm); err != nil {
		logrus.WithError(err).Fatalf("Error when mkdir %s", LogDir)
	}
}

// ConvenientErrorLogger is a thin adapter, does enhancement over logrus
type ConvenientErrorLogger struct {
	*logrus.Entry
}

// embed spread just for using ConvenientLogger.Error  but not entry.Error after chain call
func (l *ConvenientErrorLogger) WithField(key string, value interface{}) *ConvenientErrorLogger {
	return &ConvenientErrorLogger{l.Entry.WithField(key, value)}
}

func (l *ConvenientErrorLogger) WithFields(fields logrus.Fields) *ConvenientErrorLogger {
	return &ConvenientErrorLogger{l.Entry.WithFields(fields)}
}

func (l *ConvenientErrorLogger) WithObject(obj interface{}) *ConvenientErrorLogger {
	objMap := make(map[string]interface{})

	var b []byte

	b, _ = json.Marshal(&obj)

	_ = json.Unmarshal(b, &objMap)

	return l.WithFields(objMap)
}

func (l *ConvenientErrorLogger) WithError(err error) *ConvenientErrorLogger {
	return &ConvenientErrorLogger{l.Entry.WithError(err).WithField("stack", fmt.Sprintf("%+v", err))}
}

// ErrStackf is a helper to reduce log.error() then return nil, err burden
func (l *ConvenientErrorLogger) ErrStackf(err error, format string, args ...interface{}) {
	l.WithError(err).Errorf(format, args...)
}

// GetLogger return a logger with field name
func GetLogger(name string) *ConvenientErrorLogger {
	if e, ok := loggerCache[name]; ok {
		return e
	}

	l := logrus.New()

	// customFormatter := new(logrus.TextFormatter)
	// customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	// l.SetFormatter(customFormatter)

	// below can be apply on individual logger
	l.SetFormatter(&logrus.JSONFormatter{})
	// l.SetReportCaller(true)

	fileName := LogDir + strings.ReplaceAll(name+".log", "/", "_")
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		logrus.WithError(err).Fatalf("Fatal creating log file %s", fileName)
	}

	LogFileNames[name] = f

	l.SetOutput(io.MultiWriter(os.Stdout, f))

	l.SetLevel(logrus.DebugLevel)

	l.AddHook(&apmlogrus.Hook{})
	//l.AddHook(&SentryHook{})

	loggerCache[name] = &ConvenientErrorLogger{l.WithField("logger", name)}

	return loggerCache[name]
}
