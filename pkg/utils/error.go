package utils

import (
	"fmt"
	"os"

	logger "deposit-collector/pkg/logger"
)

func FailOnError(logger *logger.Logger, err error, msg string) {
	if err != nil {
		logger.Error(fmt.Sprintf("%s: %s", msg, err))
		os.Exit(1)
	}
}
