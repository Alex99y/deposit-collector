package crypto

import (
	errors "errors"
	strings "strings"

	utils "deposit-collector/pkg/utils"
)

type Path struct {
	Purpose  uint32
	CoinType uint32
	Account  uint32
	Change   uint32
	Index    uint32
}

func validateBIP44Path(path string) (*Path, error) {
	pathStruct := new(Path)
	validateLastCharacter := func(value string) error {
		lastCharacter := string(value[len(value)-1])
		if lastCharacter != "'" {
			return errors.New("value must end with '")
		}
		return nil
	}
	if path == "" {
		return nil, errors.New("path is required")
	}
	splitedPath := strings.Split(path, "/")
	if len(splitedPath) != 6 {
		return nil, errors.New("path must be in the format m/44'/60'/0'/0/0")
	}

	// Validate purpose
	if splitedPath[0] != "m" {
		return nil, errors.New("path must start with m")
	}
	purpose := splitedPath[1]
	if err := validateLastCharacter(purpose); err != nil {
		return nil, err
	}
	purpose = purpose[:len(purpose)-1]
	purposeNumber, err := utils.IsNumber(purpose)
	if err != nil {
		return nil, errors.New("purpose must be a number")
	}
	pathStruct.Purpose = uint32(purposeNumber)

	// Validate coin type
	coinType := splitedPath[2]
	if err := validateLastCharacter(coinType); err != nil {
		return nil, err
	}
	coinType = coinType[:len(coinType)-1]
	coinTypeNumber, err := utils.IsNumber(coinType)
	if err != nil {
		return nil, errors.New("coin type must be a number")
	}
	pathStruct.CoinType = uint32(coinTypeNumber)

	// Validate account
	account := splitedPath[3]
	if err := validateLastCharacter(account); err != nil {
		return nil, err
	}
	account = account[:len(account)-1]
	accountNumber, err := utils.IsNumber(account)
	if err != nil {
		return nil, errors.New("account must be a number")
	}
	pathStruct.Account = uint32(accountNumber)

	// Validate change
	change := splitedPath[4]
	changeNumber, err := utils.IsNumber(change)
	if err != nil {
		return nil, errors.New("change must be a number")
	}
	pathStruct.Change = uint32(changeNumber)

	// Validate index
	index := splitedPath[5]
	indexNumber, err := utils.IsNumber(index)
	if err != nil {
		return nil, errors.New("index must be a number")
	}
	pathStruct.Index = uint32(indexNumber)
	return pathStruct, nil
}
