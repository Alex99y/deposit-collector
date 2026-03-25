package wallet

type CryptoWallet interface {
	GetAddress() string
	SignMessage(message string) ([]byte, error)
}
