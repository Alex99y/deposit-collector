package crypto

import "fmt"

const (
	// Bitcoin
	PurposeBTCNativeSegwit = 84
	CoinTypeBTC            = 0
	CoinTypeBTCTestnet     = 1

	// Ethereum
	PurposeEVM  = 44
	CoinTypeEVM = 60

	// Solana
	PurposeSOL  = 44
	CoinTypeSOL = 501
)

type BIP44Path struct {
	Purpose  uint32
	CoinType uint32
	Account  uint32
	Change   uint32
	Index    uint32
}

func (b *BIP44Path) GeneratePath() string {
	return fmt.Sprintf(
		"m/%d'/%d'/%d'/%d/%d",
		b.Purpose,
		b.CoinType,
		b.Account,
		b.Change,
		b.Index,
	)
}

func NewBIP44(
	purpose uint32,
	coinType uint32,
	accountIndex uint32,
	changeIndex uint32,
	index uint32,
) BIP44Path {
	return BIP44Path{
		Purpose:  purpose,
		CoinType: coinType,
		Account:  accountIndex,
		Change:   changeIndex,
		Index:    index,
	}
}
