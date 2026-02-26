package utils

import (
	"errors"
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

func NewError(msg string) error {
	return errors.New(msg)
}
