package rpc

import (
	"csc/jsonrpc"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"
)

type RPCClient struct {
	jsonrpc.Client
}

func NewRPCClient(url string) RPCClient {
	return RPCClient{jsonrpc.Client{URL: url}}
}

func (c RPCClient) NextNonce(address string) (uint64, error) {
	res, err := c.Do("eth_getTransactionCount", address, "latest")
	if err != nil {
		return 0, err
	}
	var hexN string
	if err = res.Result(&hexN); err != nil {
		return 0, err
	}
	return strconv.ParseUint(hexN[2:], 16, 64)
}

func (c RPCClient) EstimateGas(tx map[string]string) (gas uint64, err error) {
	res, err := c.Do("eth_estimateGas", tx)
	if err != nil {
		return
	}
	var hexGas string
	if err = res.Result(&hexGas); err != nil {
		return
	}
	gas, err = strconv.ParseUint(hexGas[2:], 16, 64)
	if err != nil {
		return
	}
	gas = gas * 4 / 3 // estimate gas maybe not enough
	return
}

func (c RPCClient) WaitTxConfirmed(txid string) error {
	var confirmed bool
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second * 3)
		res, err := c.Do("eth_getTransactionByHash", txid)
		if err != nil {
			continue
		}
		var ret struct{ BlockNumber string }
		if err = res.Result(&ret); err != nil {
			continue
		}
		if ret.BlockNumber != "" {
			confirmed = true
			break
		}
	}
	if !confirmed {
		return errors.New("timeout")
	}
	for i := 0; i < 3; i++ {
		res, err := c.Do("eth_getTransactionReceipt", txid)
		if err != nil {
			continue
		}
		var ret struct{ Status string }
		if err = res.Result(&ret); err != nil {
			continue
		}
		if ret.Status == "0x1" {
			return nil
		}
	}
	return errors.New("failed")
}

func (c RPCClient) TokenBalance(address string, contractAddress string) (*big.Int, error) {
	res, err := c.Do("eth_call", map[string]string{
		"from": "0x0000000000000000000000000000000000000000",
		"to":   contractAddress,
		"data": "0x70a08231000000000000000000000000" + address[2:],
	}, "latest")
	if err != nil {
		return nil, err
	}
	var hexAmount string
	if err = res.Result(&hexAmount); err != nil {
		return nil, err
	}
	v, ok := big.NewInt(0).SetString(hexAmount[2:], 16)
	if !ok {
		return nil, fmt.Errorf("parse %s failed", hexAmount)
	}
	return v, nil
}

func (c RPCClient) AddressBalance(address string) (*big.Int, error) {
	res, err := c.Do("eth_getBalance", address, "latest")
	if err != nil {
		return nil, err
	}
	var hexAmount string
	if err = res.Result(&hexAmount); err != nil {
		return nil, err
	}
	v, ok := big.NewInt(0).SetString(hexAmount[2:], 16)
	if !ok {
		return nil, fmt.Errorf("parse %s failed", hexAmount)
	}
	return v, nil
}
