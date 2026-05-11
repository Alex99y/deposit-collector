package walletservices

import (
	errors "errors"

	evm_utils "deposit-collector/pkg/crypto/evm"
	wallet "deposit-collector/pkg/crypto/wallet"

	common "github.com/ethereum/go-ethereum/common"
	crypto "github.com/ethereum/go-ethereum/crypto"
	types "github.com/ethereum/go-ethereum/core/types"
)

type EVMSigner struct {
	Wallet wallet.EvmWallet
	Txs    []*evm_utils.UnsignedEIP1559Tx
}

// SignEVMTransactions signs one or more unsigned EIP-1559 transactions.
// It returns the signed transactions in the same order as provided.
func SignEVMTransactions(
	signers []EVMSigner,
) ([]*types.Transaction, error) {
	if len(signers) == 0 {
		return nil, errors.New("no evm signers provided")
	}

	signedTxs := make([]*types.Transaction, 0)

	for _, signer := range signers {
		privKey, err := crypto.HexToECDSA(signer.Wallet.PrivateKey)
		if err != nil {
			return nil, err
		}

		signerAddr := common.HexToAddress(signer.Wallet.Address)

		for _, unsignedTx := range signer.Txs {
			if unsignedTx == nil || unsignedTx.Tx == nil {
				return nil, errors.New("invalid unsigned tx")
			}
			if signerAddr != unsignedTx.From {
				return nil, errors.New("signer address does not match unsigned tx from")
			}

			londonSigner := types.NewLondonSigner(unsignedTx.Tx.ChainID)
			signedTx, err := types.SignTx(types.NewTx(unsignedTx.Tx), londonSigner, privKey)
			if err != nil {
				return nil, err
			}
			signedTxs = append(signedTxs, signedTx)
		}
	}

	return signedTxs, nil
}

