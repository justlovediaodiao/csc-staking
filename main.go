package main

import (
	"csc/provider"
	"csc/rpc"
	"flag"
	"fmt"
	"math/big"
	"strings"
)

func main() {
	const chainID = 52
	const contractAddress = "0x0000000000000000000000000000000000001000"

	var validator string
	flag.StringVar(&validator, "validator", "0x62f7f2f03dc042baf765003ff0f4011720a20596", "validator address")
	flag.Parse()

	hexData := fmt.Sprintf("0x26476204%064s", strings.ToLower(validator[2:]))
	gasPrice, _ := new(big.Int).SetString("500000000000", 10) // 500GWei
	var gas uint64 = 183817
	base, _ := new(big.Int).SetString("1000000000000000000", 10) // 1e18

	p, err := provider.NewProvider(chainID)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer p.Close()

	c := rpc.NewRPCClient("https://rpc.coinex.net/")
	amount, err := c.AddressBalance(p.Account())
	if err != nil {
		fmt.Println(err)
		return
	}
	// align to int and retain 1 CET
	amount.Div(amount, base).Mul(amount, base)
	amount.Sub(amount, base)
	// min staking is 1000 CET
	if amount.Cmp(new(big.Int).Mul(base, big.NewInt(1000))) < 0 {
		fmt.Println("balance is less than 1000 CET")
		return
	}

	nonce, err := c.NextNonce(p.Account())
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("staking %s CET...\n", new(big.Int).Div(amount, base))

	txid, err := p.SendTransaction(nonce, contractAddress, amount, gas, gasPrice, hexData)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("txid: %s\n", txid)
}
