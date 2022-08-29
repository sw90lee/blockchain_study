package block

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/sw90lee/blockchain_study/utils"
)

// 앞 0이 3개가나오면 작업증명완료 난이도 설정
const (
	MINING_DIFFICUITY = 2
	MINING_SENDER     = "THE BLOCKCHAIN"
	MINING_REWORD     = 1.0
)

// Block 구조
type Block struct {
	timestamp    int64
	nonce        int
	previousHash [32]byteTransactions
	transactions []*Transactions
}

func NewBlock(nonce int, previousHash [32]byte, transactions []*Transactions) *Block {
	b := new(Block)
	b.timestamp = time.Now().UnixNano()
	b.nonce = nonce
	b.previousHash = previousHash
	b.transactions = transactions
	return b
}

func (b *Block) Print() {
	fmt.Printf("timestamp       %d\n", b.timestamp)
	fmt.Printf("nonce           %d\n", b.nonce)
	fmt.Printf("previous_hash   %x\n", b.previousHash)
	fmt.Printf("transactions    %v\n", b.transactions)
	for _, t := range b.transactions {
		t.Print()
	}
}

func (b *Block) Hash() [32]byte {
	m, _ := json.Marshal(b)
	return sha256.Sum256([]byte(m))
}

func (b *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Timestamp    int64           `json:"timestamp"`
		Nonce        int             `json:"nonce"`
		PreviousHash [32]byte        `json:"previous_hash"`
		Transactions []*Transactions `json:"transactions"`
	}{
		Timestamp:    b.timestamp,
		Nonce:        b.nonce,
		PreviousHash: b.previousHash,
		Transactions: b.transactions,
	})
}

// Transaction 구조
type Transactions struct {
	senderBlockchainAddres     string
	recipientBlockchainAddress string
	value                      float32
}

func NewTransation(sender, recipient string, value float32) *Transactions {
	return &Transactions{sender, recipient, value}
}

func (t *Transactions) Print() {
	fmt.Printf("%s\n", strings.Repeat("-", 40))
	fmt.Printf(" sender_blockchain_address	%s\n", t.senderBlockchainAddres)
	fmt.Printf(" recipient_blockchain_address	%s\n", t.recipientBlockchainAddress)
	fmt.Printf(" value	                        %.1f\n", t.value)
}

func (t *Transactions) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Sender    string  `json:"sender_blockchain_address"`
		Recipient string  `json:"recipient_blockchain_address"`
		Value     float32 `json:"value"`
	}{
		Sender:    t.recipientBlockchainAddress,
		Recipient: t.recipientBlockchainAddress,
		Value:     t.value,
	})
}

// BlockChain 구조
type Blockchain struct {
	transactionPool   []*Transactions
	chain             []*Block
	blockchainAddress string
}

func NewBlockchain(blockchainAddress string) *Blockchain {
	b := &Block{}
	bc := new(Blockchain)
	bc.blockchainAddress = blockchainAddress
	bc.CreateBlock(0, b.Hash())
	return bc
}

func (bc *Blockchain) CreateBlock(nonce int, previousHash [32]byte) *Block {
	b := NewBlock(nonce, previousHash, bc.transactionPool)
	bc.chain = append(bc.chain, b)
	bc.transactionPool = []*Transactions{}
	return b
}

func (bc *Blockchain) LastBlock() *Block {
	return bc.chain[len(bc.chain)-1]
}

func (bc *Blockchain) Print() {
	for i, block := range bc.chain {
		fmt.Printf("%s Chain %d %s\n", strings.Repeat("=", 25), i, strings.Repeat("=", 25))
		block.Print()
	}
	fmt.Printf("%s\n", strings.Repeat("*", 25))
}

func (bc *Blockchain) CopyTransacionPool() []*Transactions {
	transactions := make([]*Transactions, 0)
	for _, t := range bc.transactionPool {
		transactions = append(transactions, NewTransation(t.senderBlockchainAddres, t.recipientBlockchainAddress, t.value))
	}
	return transactions
}

// 합의 알고리즘
func (bc *Blockchain) ValidProof(nonce int, previousHash [32]byte, tranTransactions []*Transactions, difficulty int) bool {
	// 0이 나오는 만큼 난이도 설정
	zeros := strings.Repeat("0", difficulty)
	guessBlock := Block{0, nonce, previousHash, tranTransactions}
	guessHashstr := fmt.Sprintf("%x", guessBlock.Hash())
	fmt.Println(guessHashstr)
	return guessHashstr[:difficulty] == zeros
}

func (bc *Blockchain) ProofOfWork() int {
	transactions := bc.CopyTransacionPool()
	previousHash := bc.LastBlock().Hash()
	nonce := 0
	for !bc.ValidProof(nonce, previousHash, transactions, MINING_DIFFICUITY) {
		nonce += 1
	}
	return nonce
}

/////////////////////////////////////
// Mining
func (bc *Blockchain) Mining() bool {
	bc.AddTransaction(MINING_SENDER, bc.blockchainAddress, MINING_REWORD, nil, nil)
	nonce := bc.ProofOfWork()
	previousHash := bc.LastBlock().Hash()
	bc.CreateBlock(nonce, previousHash)
	log.Println("action = mining, status = success")
	return true
}

// 코인 총량 찾기
func (bc *Blockchain) CaculateTotalAmount(blockchainAddress string) float32 {
	var totalAmount float32 = 0.0
	for _, b := range bc.chain {
		for _, t := range b.transactions {
			value := t.value
			if blockchainAddress == t.recipientBlockchainAddress {
				totalAmount += value
			}

			if blockchainAddress == t.senderBlockchainAddres {
				totalAmount -= value
			}
		}
	}
	return totalAmount
}

func (bc *Blockchain) AddTransaction(sender string, recipient string, value float32, senderPublicKey *ecdsa.PublicKey, s *utils.Signature) bool {
	t := NewTransation(sender, recipient, value)
	bc.transactionPool = append(bc.transactionPool, t)

	if sender == MINING_SENDER {
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	}

	if bc.VerifyTransationSignature(senderPublicKey, s, t) {
		if bc.CaculateTotalAmount(sender) < value {
			log.Println("ERROR: 지갑에 충분한 balance가 없습니다.")
		}
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	} else {
		log.Println("ERROR: verify Transecation")
	}
	return false
}

func (bc *Blockchain) VerifyTransationSignature(senderPublicKey *ecdsa.PublicKey, s *utils.Signature, t *Transactions) bool {
	m, _ := json.Marshal(t)
	h := sha256.Sum256(m)
	return ecdsa.Verify(senderPublicKey, h[:], s.R, s.S)
}


type TransactionRequest strcut {
	SenderBlockchainAddress *string `json:"sender_blockchain_address"`
	RecipientBlockchainAddress *string `json:"recipient_blockchain_address`
	SenderPublicKey *string `json:"sender_public_key"`
	Value *float32 `json:"value"`
	Signature *string `json:"signature"`
}

func (tr *TransactionRequest) Validate() bool {
	if tr.SenderBlockchainAddress == nil ||
		tr.RecipientBlockchainAddress == nil ||
		tr.SenderPublicKey == nil ||
		tr.Value == nil ||
		tr.Signature == nil {
			return false
		}
		return true
}