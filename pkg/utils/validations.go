package utils

import "strconv"

func IsNumber(value string) (int, error) {
	number, err := strconv.Atoi(value)
	return number, err
}
