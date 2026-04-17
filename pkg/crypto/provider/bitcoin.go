package provider

import (
	"bufio"
	context "context"
	json "encoding/json"
	"errors"
	"net"
	"strings"
	"time"

	btc_utils "deposit-collector/pkg/crypto/btc"
	logger "deposit-collector/pkg/logger"
	utils "deposit-collector/pkg/utils"
)

type ElectrumClient struct {
	url string
}

type Request struct {
	ID     int64         `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

type Response struct {
	ID     int64           `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  interface{}     `json:"error"`
}

func (c *ElectrumClient) Request(req Request) (*Response, error) {
	splittedUrl := strings.Split(c.url, "//")
	conn, err := net.DialTimeout("tcp", splittedUrl[1], 5*time.Second)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	payload, err := json.Marshal(req)

	_, err = conn.Write(append(payload, '\n'))
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)

	line, err := reader.ReadBytes('\n')

	if err != nil {
		return nil, err
	}

	var res Response

	if err = json.Unmarshal(line, &res); err != nil {
		return nil, err
	}

	if res.Error != nil {
		// @todo check this
		return nil, errors.New("Something went wrong")
	}

	return &res, nil
}

func NewElectrumClient(url string) (*ElectrumClient, error) {
	return &ElectrumClient{
		url: url,
	}, nil
}

type BitcoinProvider struct {
	electrumClient   ElectrumClient
	context          context.Context
	MinConfirmations int
	Network          btc_utils.NETWORK
}

type LatestBlockResponse struct {
	Height int64  `json:"height"`
	Hex    string `json:"hex"`
}

type TxInfoResponse struct {
	TxID          string `json:"txid"`
	Hash          string `json:"hash"`
	Version       int    `json:"version"`
	Size          int    `json:"size"`
	VSize         int    `json:"vsize"`
	Weight        int    `json:"weight"`
	Locktime      int    `json:"locktime"`
	Vin           []Vin  `json:"vin"`
	Vout          []Vout `json:"vout"`
	Blockhash     string `json:"blockhash"`
	Confirmations int    `json:"confirmations"`
	Time          int64  `json:"time"`
	Blocktime     int64  `json:"blocktime"`
}

type Vin struct {
	TxID      string    `json:"txid"`
	Vout      int       `json:"vout"`
	ScriptSig ScriptSig `json:"scriptSig"`
	Sequence  int64     `json:"sequence"`
	Witness   []string  `json:"txinwitness,omitempty"`
}

type Vout struct {
	Value        float64      `json:"value"`
	N            int          `json:"n"`
	ScriptPubKey ScriptPubKey `json:"scriptPubKey"`
}

type ScriptPubKey struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex"`
	Addresses []string `json:"addresses,omitempty"`
	Type      string   `json:"type"`
	ReqSigs   int      `json:"reqSigs"`
}

type ScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}

func (p *BitcoinProvider) GetLatestBlockNumber() (uint64, error) {
	resp, err := p.electrumClient.Request(Request{
		ID:     time.Now().Unix(),
		Method: "blockchain.headers.subscribe",
		Params: []interface{}{},
	})

	if err != nil {
		return 0, err
	}

	var header LatestBlockResponse
	if err = json.Unmarshal(resp.Result, &header); err != nil {
		return 0, err
	}

	return uint64(header.Height), nil
}

func (p *BitcoinProvider) GetTxInfo(txHash string) (*TxInfoResponse, error) {
	resp, err := p.electrumClient.Request(Request{
		ID:     time.Now().Unix(),
		Method: "blockchain.transaction.get",
		Params: []interface{}{txHash, true},
	})

	if err != nil {
		return nil, err
	}

	var txInfo TxInfoResponse
	if err = json.Unmarshal(resp.Result, &txInfo); err != nil {
		return nil, err
	}

	return &txInfo, nil
}

func NewBitcoinProvider(
	url string,
	minConfirmations int,
	context context.Context,
	network btc_utils.NETWORK,
	logger *logger.Logger,
) *BitcoinProvider {
	electrumClient, err := NewElectrumClient(url)
	if err != nil {
		utils.FailOnError(logger, err, "Error creating EVM provider")
	}

	return &BitcoinProvider{
		electrumClient:   *electrumClient,
		context:          context,
		MinConfirmations: minConfirmations,
		Network:          network,
	}
}
