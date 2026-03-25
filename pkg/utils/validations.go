package utils

import "strconv"

func StringToInt(value string) (int, error) {
	number, err := strconv.Atoi(value)
	return number, err
}

func StringToInt64(value string) (int64, error) {
	number, err := strconv.ParseInt(value, 10, 64)
	return number, err
}

func IsNumber(value string) (int, error) {
	return StringToInt(value)
}
