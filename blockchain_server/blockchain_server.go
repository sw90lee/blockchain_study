package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/sw90lee/blockchain_study/block"
	"github.com/sw90lee/blockchain_study/utils"
	"github.com/sw90lee/blockchain_study/wallet"
)

var cache map[string]*block.Blockchain = make(map[string]*block.Blockchain)

type BlockchainServer struct {
	port uint16
}

func NewBlockchainServer(port uint16) *BlockchainServer {
	return &BlockchainServer{port}
}

func (bcs *BlockchainServer) Port() uint16 {
	return bcs.port
}
func (bcs *BlockchainServer) GetBlockChain() *block.Blockchain {
	bc, ok := cache["blockchain"]
	if !ok {
		minersWallet := wallet.NewWallet()
		bc = block.NewBlockchain(minersWallet.BlockchainAddress(), bcs.Port())
		cache["blockchain"] = bc
		log.Panicf("private_key : %v", minersWallet.PrivateKeyStr())
		log.Panicf("public_key : %v", minersWallet.PublicKeyStr())
		log.Panicf("blockchain_address : %v", minersWallet.BlockchainAddress())
	}
	return bc
}

func (bcs *BlockchainServer) GetChain(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		decoder := json.NewDecoder(r.Body)
		var t wallet.TransationRequest
		err := decoder.Decode(&t)
		if err != nil {
			log.Printf("ERROR: %v" ,err)
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		if !Validate() {
			log.Println("ERROR: mssing field(s)")
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		publicKey := utils.PubilcKeyFromString(*t.SenderPublicKey)
		privateKey := utils.PrivateKeyFromString(*t.SenderPrivateKey, publicKey)
		value , err := strconv.ParseFloat(*t.Value, 32)
		if err != nil {
			log.Printf("ERROR: %v" ,err)
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}

		value32 := float32(value)

		w.Header().Add("Content-type", "application/json")


		transaction := wallet.NewTransaction(privateKey, publicKey, *t.SenderBlockchainAddress, *t.RecipientBlockchainAddress, value32)
		signature := transaction.GenerateSignature()
		signatureStr := signature.String()

		bt := &block.TransactionRequest{
			t.SenderBlockchainAddress,
			t.RecipientBlockchainAddress,
			t.SenderPublicKey,
			&value32, &signatureStr,
		}
		m , _ := json.Marshal(bt)


		bc := bcs.GetBlockChain()
		m, _ := bc.MarshalJSON()
		io.WriteString(w, string(m[:]))
	default:
		log.Println("ERROR: Invalid HTTP method")
	}
}

func (bcs *BlockchainServer) Run() {
	http.HandleFunc("/", bcs.GetChain)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(bcs.Port())), nil))
}
