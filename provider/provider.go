package provider

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"runtime"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

type Provider struct {
	conn    *WalletConnection
	chainID int
	account string
	topic   string
}

func (p *Provider) Account() string {
	return p.account
}

func jsonrpc(method string, param string) string {
	id := time.Now().Unix()
	payload := `{"id":%d,"jsonrpc":"2.0","method":"%s","params":[%s]}`
	return fmt.Sprintf(payload, id, method, param)
}

func NewProvider(chainID int) (*Provider, error) {
	conn, err := NewWalletConnection()
	if err != nil {
		return nil, err
	}

	showQRCode(conn.Url())

	p := Provider{conn: conn, chainID: chainID}
	if err := p.sessionRequest(); err != nil {
		p.Close()
		return nil, err
	}
	fmt.Printf("wallet connected, address %s\n", p.account)
	return &p, nil
}

func showQRCode(url string) {
	data, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		fmt.Println("please open the url in wallet app:")
		fmt.Println(url)
		return
	}
	if err := os.WriteFile("qrcode.png", data, 0644); err != nil {
		fmt.Println("please open the url in wallet app:")
		fmt.Println(url)
		return
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("open", "qrcode.png")
	} else if runtime.GOOS == "windows" {
		cmd = exec.Command("explorer", "qrcode.png")
	} else {
		fmt.Println("please open qrcode.png and scan with wallet app...")
		return
	}
	if err := startProcess(cmd); err != nil {
		fmt.Println("please open qrcode.png and scan with wallet app...")
		return
	}
	fmt.Println("please scan the QR code with wallet app...")
}

func (p *Provider) sessionRequest() error {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return err
	}
	topic := hex.EncodeToString(buf)
	param := `{"peerId":"%s","peerMeta":{"description":"csc staking tool","url":"https://github.com/justlovediaodiao/csc-staking","icons":["https://www.coinex.org/favicon.ico"],"name":"csc-staking"},"chainId":%d}`
	param = jsonrpc("wc_sessionRequest", fmt.Sprintf(param, topic, p.chainID))
	if err := p.conn.Pub("", []byte(param)); err != nil {
		return err
	}
	if err := p.conn.Sub(topic); err != nil {
		return err
	}
	payload, err := p.conn.Receive()
	if err != nil {
		return err
	}
	var result struct {
		Result struct {
			PeerID   string   `json:"peerId"`
			Approved bool     `json:"approved"`
			Accounts []string `json:"accounts"`
		}
	}
	if err := json.Unmarshal(payload, &result); err != nil {
		return err
	}
	if !result.Result.Approved {
		return errors.New("approval denied")
	}
	p.topic = result.Result.PeerID
	p.account = result.Result.Accounts[0]
	return nil
}

func (p *Provider) SendTransaction(nonce uint64, to string, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data string) (string, error) {
	param := `{"from":"%s","to":"%s","data":"%s","gas":"0x%x","gasPrice":"0x%x","value":"0x%x","nonce":"0x%x"}`
	param = jsonrpc("eth_sendTransaction", fmt.Sprintf(param, p.account, to, data, gasLimit, gasPrice, amount, nonce))
	if err := p.conn.Pub(p.topic, []byte(param)); err != nil {
		return "", err
	}
	payload, err := p.conn.Receive()
	if err != nil {
		return "", err
	}
	var result struct {
		Result string
	}
	if err := json.Unmarshal(payload, &result); err != nil {
		return "", err
	}
	if result.Result == "" {
		return "", errors.New("send transaction failed")
	}
	return result.Result, nil
}

func (p *Provider) Close() error {
	if p.conn != nil {
		err := p.conn.Close()
		p.conn = nil
		return err
	}
	return nil
}
