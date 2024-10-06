package app

import (
	"github.com/sirupsen/logrus"
	"os"
)

type CustomConsoleFormatter struct{}

func (f *CustomConsoleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	output := []byte(entry.Time.Format("2006/01/02 15:04:05") + " " + entry.Message + "\n")
	return output, nil
}

func SetLevel(lvl string, formatter string) {
	level, err := logrus.ParseLevel(lvl)
	if err != nil {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(level)
	}

	var format logrus.Formatter
	if formatter == "console" {
		format = &CustomConsoleFormatter{}
	} else {
		format = &logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		}
	}

	logrus.SetFormatter(format)
	logrus.SetOutput(os.Stdout)
}
