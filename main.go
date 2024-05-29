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
	//合约名称-------------------------------------------------(修改处)
	contract_name := "SimpleStorage"
	// 连接到Ganache
	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	//获取chainID
	if err != nil {
		log.Fatalf("Failed to retrieve network ID: %v", err)
	}
	//获得第一个账户的私钥
	privatekey := utils.GetENV("PRIVATE_KEY_1")
	fmt.Println(privatekey)

	if err != nil {
		panic(err)
	}
	// 构建新交易(gasLimit--gas限制，value--交易值)
	auth1 := utils.Transact(client, privatekey, big.NewInt(0))

	//部署合约（服务器，区块链ID，合约名称，私钥）
	address, tx0 := utils.Deploy(client, contract_name, auth1)
	//获取部署合约的gas值
	receipt, err := bind.WaitMined(context.Background(), client, tx0)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("Deploy Gas used: %d\n", receipt.GasUsed)

	//构建调用合约实体
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

	// 调用get方法获取值
	storedValue, _ := Contract.Get(&bind.CallOpts{})
	fmt.Printf("Stored value: %s\n", storedValue)
}
