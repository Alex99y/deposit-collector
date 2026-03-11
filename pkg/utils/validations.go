package utils

import "strconv"

func StringToInt(value string) (int, error) {
	number, err := strconv.Atoi(value)
	return number, err
}

func IsNumber(value string) (int, error) {
	return StringToInt(value)
}
