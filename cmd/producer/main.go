package main

import (
	"WB_ZeroProject/internal/app"
	"github.com/sirupsen/logrus"
)

func main() {

	err := app.RunProducer()
	if err != nil {
		logrus.Errorf("app.RunProducer%s", err.Error())
	}

}
