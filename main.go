package main

import (
	"fmt"
	"log"

	"github.com/sw90lee/blockchain_study/wallet"
)

func init() {
	log.SetPrefix("Blockchain: ")
}

func main() {
	w := wallet.NewWallet()
	fmt.Println(w.PrivateKey())
	fmt.Println(w.PublicKey())
	fmt.Println(w.PrivateKeyStr())
	fmt.Println(w.PublicKeyStr())
}
