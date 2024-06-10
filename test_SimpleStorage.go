package main

import (
	"context"
	//"dttp/compile/contract"
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
	storedValue, _ := Contract.Get(&bind.CallOpts{})
	receipt, _ = bind.WaitMined(context.Background(), client, tx)
	fmt.Printf("Set() Gas used: %d\n", receipt.GasUsed)
	fmt.Printf("Stored value: %s\n", storedValue)

	// Construct a G1Point type
	g1Point := contract.SimpleStorageG1Point{
		X: big.NewInt(123),
		Y: big.NewInt(456),
	}

	auth3 := utils.Transact(client, privatekey, big.NewInt(0))
	//invoke SetG1Point
	tx3, _ := Contract.SetG1Point(auth3, g1Point)
	receipt3, _ := bind.WaitMined(context.Background(), client, tx3)
	G1PointValue, _ := Contract.GetG1Point(&bind.CallOpts{})
	fmt.Printf("getG1Point() Gas used: %d\n", receipt3.GasUsed)
	fmt.Printf("G1Point value: %v\n", G1PointValue)
}
