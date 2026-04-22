package walletservices

import (
	"bytes"
	hex "encoding/hex"
	errors "errors"

	btc_utils "deposit-collector/pkg/crypto/btc"
	wallet "deposit-collector/pkg/crypto/wallet"

	"github.com/btcsuite/btcd/wire"
)

type BitcoinSigner struct {
	Wallet wallet.BitcoinWallet
	Inputs []wallet.BitcoinTransactionInput
}

func SignTransactionWithInputs(
	network btc_utils.NETWORK,
	unsignedTxHex string,
	signers []BitcoinSigner,
) (string, error) {
	if len(signers) == 0 {
		return "", errors.New("no bitcoin signers provided")
	}

	var tx wire.MsgTx

	txBytes, err := hex.DecodeString(unsignedTxHex)
	if err != nil {
		return "", err
	}

	if err := tx.Deserialize(bytes.NewReader(txBytes)); err != nil {
		return "", err
	}

	for _, signer := range signers {
		signedTx, err := signer.Wallet.SignTransactionInputs(
			tx,
			network,
			signer.Inputs,
		)
		if err != nil {
			return "", err
		}
		tx = *signedTx
	}
	var signedTxBuffer bytes.Buffer
	if err := tx.Serialize(&signedTxBuffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(signedTxBuffer.Bytes()), nil
}
