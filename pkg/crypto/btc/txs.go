package btc_utils

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

const (
	DefaultTxVersion     = 2
	DefaultInputSequence = wire.MaxTxInSequenceNum
	DefaultDustThreshold = int64(546)
	// version (4) + locktime (4) + marker/flag (2) + varints ~= 11 vbytes.
	p2wpkhBaseTxVBytes = 11
	// Typical vbytes for one P2WPKH output.
	p2wpkhOutputVBytes = 31
)

type UTXOInput struct {
	TxHash     string
	Vout       uint32
	AmountSats int64
	Sequence   uint32
}

type TxOutput struct {
	Address    string
	AmountSats int64
}

type CreateTransactionRequest struct {
	Network       NETWORK
	Inputs        []UTXOInput
	Outputs       []TxOutput
	FeeSats       int64
	ChangeAddress string
	LockTime      uint32
	Version       int32
	DustThreshold int64
}

type CreateTransactionResult struct {
	RawHex          string
	TotalInputSats  int64
	TotalOutputSats int64
	ChangeSats      int64
	FeeSats         int64
}

func CreateTransaction(
	req CreateTransactionRequest,
) (*CreateTransactionResult, error) {
	if len(req.Inputs) == 0 {
		return nil, errors.New("at least one input is required")
	}
	if len(req.Outputs) == 0 {
		return nil, errors.New("at least one output is required")
	}
	if req.FeeSats < 0 {
		return nil, errors.New("fee cannot be negative")
	}

	params := GetNetParamsByNetwork(req.Network)
	if params == nil {
		return nil, errors.New("invalid bitcoin network")
	}

	txVersion := req.Version
	if txVersion == 0 {
		txVersion = DefaultTxVersion
	}

	dustThreshold := req.DustThreshold
	if dustThreshold == 0 {
		dustThreshold = DefaultDustThreshold
	}

	tx := wire.NewMsgTx(txVersion)
	tx.LockTime = req.LockTime

	totalInput := int64(0)
	for _, input := range req.Inputs {
		if input.AmountSats <= 0 {
			return nil, fmt.Errorf(
				"invalid input amount for %s:%d", input.TxHash, input.Vout,
			)
		}

		hash, err := chainhash.NewHashFromStr(input.TxHash)
		if err != nil {
			return nil, fmt.Errorf("invalid input tx hash %s: %w", input.TxHash, err)
		}

		outPoint := wire.NewOutPoint(hash, input.Vout)
		sequence := input.Sequence
		if sequence == 0 {
			sequence = DefaultInputSequence
		}

		tx.AddTxIn(wire.NewTxIn(outPoint, nil, nil))
		tx.TxIn[len(tx.TxIn)-1].Sequence = sequence
		totalInput += input.AmountSats
	}

	totalRequestedOutputs := int64(0)
	for _, output := range req.Outputs {
		if output.AmountSats <= 0 {
			return nil, fmt.Errorf(
				"invalid output amount for address %s: %w",
				output.Address,
				errors.New("amount is less than or equal to 0"),
			)
		}
		addr, err := btcutil.DecodeAddress(output.Address, params)
		if err != nil {
			return nil, fmt.Errorf(
				"invalid output address %s: %w", output.Address, err,
			)
		}
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to build output script for %s: %w", output.Address, err,
			)
		}
		tx.AddTxOut(wire.NewTxOut(output.AmountSats, pkScript))
		totalRequestedOutputs += output.AmountSats
	}

	requiredAmount := totalRequestedOutputs + req.FeeSats
	if totalInput < requiredAmount {
		return nil, fmt.Errorf(
			"insufficient funds: inputs=%d outputs=%d fee=%d",
			totalInput,
			totalRequestedOutputs,
			req.FeeSats,
		)
	}

	changeAmount := totalInput - requiredAmount
	if changeAmount > 0 {
		if req.ChangeAddress == "" {
			if changeAmount <= dustThreshold {
				req.FeeSats += changeAmount
				changeAmount = 0
			} else {
				return nil, errors.New(
					"change address required when change is above dust threshold",
				)
			}
		} else {
			changeAddress, err := btcutil.DecodeAddress(req.ChangeAddress, params)
			if err != nil {
				return nil, fmt.Errorf(
					"invalid change address %s: %w", req.ChangeAddress, err,
				)
			}
			changeScript, err := txscript.PayToAddrScript(changeAddress)
			if err != nil {
				return nil, fmt.Errorf("failed to build change script: %w", err)
			}
			tx.AddTxOut(wire.NewTxOut(changeAmount, changeScript))
		}
	}

	var raw bytes.Buffer
	if err := tx.Serialize(&raw); err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	return &CreateTransactionResult{
		RawHex:          fmt.Sprintf("%x", raw.Bytes()),
		TotalInputSats:  totalInput,
		TotalOutputSats: totalRequestedOutputs + changeAmount,
		ChangeSats:      changeAmount,
		FeeSats:         req.FeeSats,
	}, nil
}

type CalculateFeeRequest struct {
	InputCount         int
	OutputCount        int
	SignaturesPerInput int
	FeeRateSatPerVByte float64
	IncludeChange      bool
}

func CalculateFee(req CalculateFeeRequest) (int64, error) {
	if req.InputCount <= 0 {
		return 0, errors.New("input count must be greater than zero")
	}
	if req.OutputCount <= 0 {
		return 0, errors.New("output count must be greater than zero")
	}
	if req.SignaturesPerInput <= 0 {
		return 0, errors.New("signatures per input must be greater than zero")
	}
	if req.FeeRateSatPerVByte <= 0 {
		return 0, errors.New("fee rate must be greater than zero")
	}

	totalOutputs := req.OutputCount
	if req.IncludeChange {
		totalOutputs++
	}

	// P2WPKH input virtual size:
	// base(non-witness): outpoint(36) + scriptSigLen(1) + sequence(4) = 41 bytes
	// witness weight for one signature + one compressed pubkey:
	// count(1) + sigPushLen(1) + sig(72) + pubKeyPushLen(1) + pubKey(33) = 108 wu
	// => 41 + 108/4 = 68 vbytes per input
	p2wpkhInputVBytes := 41 + int(math.Ceil(float64(108*req.SignaturesPerInput)/4))

	txVBytes := p2wpkhBaseTxVBytes +
		(req.InputCount * p2wpkhInputVBytes) +
		(totalOutputs * p2wpkhOutputVBytes)

	fee := int64(math.Ceil(float64(txVBytes) * req.FeeRateSatPerVByte))
	if fee < 1 {
		fee = 1
	}
	return fee, nil
}

// ElectrumListUnspentEntry matches Electrum blockchain.scripthash.listunspent elements.
type ElectrumListUnspentEntry struct {
	TxHash string
	TxPos  int
	Value  int64
}

// SelectBTCInputsAndFeeForWithdrawals picks P2WPKH inputs and fee so that
// change goes to signerAddress when above dust; otherwise the remainder is
// absorbed into the fee (no dusty change output).
func SelectBTCInputsAndFeeForWithdrawals(
	electrum []ElectrumListUnspentEntry,
	outputs []TxOutput,
	satPerVByte float64,
	signerAddress string,
	totalWithdrawSats int64,
) ([]UTXOInput, int64, string, error) {
	inputs := make([]UTXOInput, 0, len(electrum))
	for _, u := range electrum {
		if u.Value <= 0 || u.TxHash == "" || u.TxPos < 0 {
			continue
		}
		inputs = append(inputs, UTXOInput{
			TxHash:     u.TxHash,
			Vout:       uint32(u.TxPos),
			AmountSats: u.Value,
		})
	}
	if len(inputs) == 0 {
		return nil, 0, "", errors.New("no spendable UTXOs")
	}
	sort.Slice(inputs, func(i, j int) bool {
		return inputs[i].AmountSats > inputs[j].AmountSats
	})

	nOut := len(outputs)
	selected := make([]UTXOInput, 0, len(inputs))
	var totalIn int64

	for _, u := range inputs {
		selected = append(selected, u)
		totalIn += u.AmountSats

		feeWithChange, err := CalculateFee(CalculateFeeRequest{
			InputCount:         len(selected),
			OutputCount:        nOut,
			SignaturesPerInput: 1,
			FeeRateSatPerVByte: satPerVByte,
			IncludeChange:      true,
		})
		if err != nil {
			return nil, 0, "", err
		}

		if totalIn >= totalWithdrawSats+feeWithChange {
			change := totalIn - totalWithdrawSats - feeWithChange
			if change > 0 && change < DefaultDustThreshold {
				feeNoChange, err := CalculateFee(CalculateFeeRequest{
					InputCount:         len(selected),
					OutputCount:        nOut,
					SignaturesPerInput: 1,
					FeeRateSatPerVByte: satPerVByte,
					IncludeChange:      false,
				})
				if err != nil {
					return nil, 0, "", err
				}
				remainderFee := totalIn - totalWithdrawSats
				if remainderFee < feeNoChange {
					continue
				}
				return selected, remainderFee, "", nil
			}
			return selected, feeWithChange, signerAddress, nil
		}
	}

	return nil, 0, "", errors.New("insufficient UTXO value for withdrawals and network fee")
}

// BitcoinTxIDFromSerializedHex returns the transaction id (BIP141 txid: double SHA-256
// of the non-witness serialization). It matches explorer / Electrum tx hashes for
// both legacy and witness transactions.
func BitcoinTxIDFromSerializedHex(txHex string) (string, error) {
	raw, err := hex.DecodeString(txHex)
	if err != nil {
		return "", fmt.Errorf("decode tx hex: %w", err)
	}
	var tx wire.MsgTx
	if err := tx.Deserialize(bytes.NewReader(raw)); err != nil {
		return "", fmt.Errorf("deserialize bitcoin tx: %w", err)
	}
	return tx.TxID(), nil
}
