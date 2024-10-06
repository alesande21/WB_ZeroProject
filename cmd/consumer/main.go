package main

import (
	"WB_ZeroProject/internal/app"
	"github.com/sirupsen/logrus"
)

func main() {

	err := app.RunConsumer()
	if err != nil {
		logrus.Errorf("app.RunConsumer%s", err.Error())
	}

}
