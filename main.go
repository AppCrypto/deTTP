package main

import (
	"context"
	"dttp/compile/contract"
	"dttp/utils"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	contract_name := "SimpleStorage"
	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	privatekey := utils.GetENV("PRIVATE_KEY_1")
	fmt.Println(privatekey)

	auth1 := utils.Transact(client, privatekey, big.NewInt(0))

	address, tx0 := utils.Deploy(client, contract_name, auth1)

	receipt, err := bind.WaitMined(context.Background(), client, tx0)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("Deploy Gas used: %d\n", receipt.GasUsed)

	Contract, err := contract.NewContract(common.HexToAddress(address.Hex()), client)
	if err != nil {
		fmt.Println(err)
	}

	auth2 := utils.Transact(client, privatekey, big.NewInt(0))
	//invoke Set
	tx, _ := Contract.Set(auth2, "dadasda")

	receipt, _ = bind.WaitMined(context.Background(), client, tx)
	fmt.Printf("Set() Gas used: %d\n", receipt.GasUsed)

	// fmt.Printf("Set transaction hash: %s\n", tx.Hash().Hex())

	storedValue, _ := Contract.Get(&bind.CallOpts{})
	fmt.Printf("Stored value: %s\n", storedValue)
}
