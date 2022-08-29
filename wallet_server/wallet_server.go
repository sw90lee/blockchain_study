package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"path"
	"strconv"

	"github.com/sw90lee/blockchain_study/block"
	"github.com/sw90lee/blockchain_study/utils"
	"github.com/sw90lee/blockchain_study/wallet"
)

const tempdir = "./template"

type WalletServer struct {
	port    uint16
	gateway string // 주소를 얻기위한 gateway
}

func NewWalletServer(port uint16, gateway string) *WalletServer {
	return &WalletServer{port, gateway}
}

func (ws *WalletServer) Port() uint16 {
	return ws.port
}

func (ws *WalletServer) Gateway() string {
	return ws.gateway
}

func (ws *WalletServer) Index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(path.Join(tempdir) + "/index.html")
	if err != nil {
		log.Fatal(err)
	}
	t.Execute(w, "")
}
func (ws *WalletServer) Wallet(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	mywallet := wallet.NewWallet()
	m, _ := mywallet.MarshalJSON()
	io.WriteString(w, string(m[:]))
}

func (ws *WalletServer) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		decoder := json.NewDecoder(r.Body)
		var t wallet.TransationRequest
		err := decoder.Decode(&t)

		if err != nil {
			log.Printf("ERROR: %v", err)
			io.WriteString(w, string(utils.JsonStatus("Failed")))
			return
		}

		if !t.Validate() {
			log.Println("ERROR: missing filed(s)")
			io.WriteString(w, string(utils.JsonStatus("failed")))
			return
		}

		publicKey := utils.PubilcKeyFromString(*t.SenderPublicKey)
		privateKey := utils.PrivateKeyFromString(*t.SenderPrivateKey, publicKey)
		value, err := strconv.ParseFloat(*t.Value, 32)
		if err != nil {
			log.Println("ERROR: parser error")
			io.WriteString(w, string(utils.JsonStatus("failed")))
			return
		}

		value32 := float32(value)

		w.Header().Add("Content-Type", "application/json")

		transaction := wallet.NewTransaction(privateKey, publicKey, *t.SenderBlockchainAddress, *t.RecipientBlockchainAddress, value32)
		signature := transaction.GenerateSignature()
		signatureStr := signature.String()

		bt := &block.TransactionRequest{
			t.SenderBlockchainAddress,
			t.RecipientBlockchainAddress,
			t.SenderPublicKey,
			&value32,
			&signatureStr,
		}

		m, _ := json.Marshal(bt)
		buf := bytes.NewBuffer(m)

		resq, err := http.Post(ws.Gateway()+"/transactions", "application/json", buf)
		if resq.StatusCode == 201 {
			io.WriteString(w, string(utils.JsonStatus("success")))
			return
		}
		io.WriteString(w, string(utils.JsonStatus("failed")))
	}
}

func (ws *WalletServer) Run() {
	http.HandleFunc("/", ws.Index)
	http.HandleFunc("/wallet", ws.Wallet)
	http.HandleFunc("/transcation", ws.CreateTransaction)
	log.Fatal(http.ListenAndServe("127.0.0.1:"+strconv.Itoa(int(ws.Port())), nil))
}
