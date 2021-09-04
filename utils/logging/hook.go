package logging

import (
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type SentryHook struct {
	LogLevels []logrus.Level
}

func (sh *SentryHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}

	sentry.CaptureException(errors.New(line))

	return nil
}

func (sh *SentryHook) Levels() []logrus.Level {
	levels := sh.LogLevels
	if levels == nil {
		levels = []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
		}
	}

	return levels
}
