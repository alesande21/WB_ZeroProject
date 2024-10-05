package app

import (
	"github.com/sirupsen/logrus"
	"os"
)

func SetLevel(lvl string) {
	level, err := logrus.ParseLevel(lvl)
	if err != nil {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(level)
	}

	format := logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	}
	logrus.SetFormatter(&format)

	logrus.SetOutput(os.Stdout)
}
