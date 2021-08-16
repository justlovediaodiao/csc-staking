package provider

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

type wsMsg struct {
	Topic   string `json:"topic"`
	Type    string `json:"type"`
	Payload string `json:"payload"`
	Silent  bool   `json:"silent"`
}

type WalletConnection struct {
	conn   *websocket.Conn
	bridge *os.Process
	server string
	topic  string
	key    []byte
}

func NewWalletConnection() (*WalletConnection, error) {
	buf := make([]byte, 48)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	c := WalletConnection{
		server: lanIP() + ":7000",
		topic:  hex.EncodeToString(buf[:16]),
		key:    buf[16:],
	}
	if err := c.connect(); err != nil {
		c.Close()
		return nil, err
	}
	return &c, nil
}

func (w *WalletConnection) Url() string {
	params := url.Values{}
	params.Add("bridge", "http://"+w.server)
	params.Add("key", hex.EncodeToString(w.key))
	return fmt.Sprintf("wc:%s@1?%s", w.topic, params.Encode())
}

func (w *WalletConnection) connect() error {
	process, err := startBridge(w.server)
	if err != nil {
		return err
	}
	w.bridge = process
	time.Sleep(time.Millisecond * 500)
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+w.server, nil)
	if err != nil {
		return err
	}
	w.conn = conn
	return nil
}

func (w *WalletConnection) Pub(topic string, msg []byte) error {
	if topic == "" {
		topic = w.topic
	}
	payload, err := Encrypt(msg, w.key)
	if err != nil {
		return err
	}
	m := wsMsg{
		Topic:   topic,
		Type:    "pub",
		Payload: payload.Marshal(),
		Silent:  true,
	}
	return w.conn.WriteJSON(m)
}

func (w *WalletConnection) Sub(topic string) error {
	if topic == "" {
		topic = w.topic
	}
	m := wsMsg{
		Topic:  topic,
		Type:   "sub",
		Silent: true,
	}
	return w.conn.WriteJSON(m)
}

func (w *WalletConnection) Receive() ([]byte, error) {
	var msg wsMsg
	if err := w.conn.ReadJSON(&msg); err != nil {
		return nil, err
	}
	var payload Payload
	if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
		return nil, err
	}
	return Decrypt(&payload, w.key)
}

func (w *WalletConnection) Close() error {
	if w.conn != nil {
		w.conn.Close()
		w.conn = nil
	}
	if w.bridge != nil {
		w.bridge.Kill()
		w.bridge = nil
	}
	return nil
}
