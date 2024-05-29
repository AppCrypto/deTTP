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
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatalf("Failed to retrieve network ID: %v", err)
	}
	utils.SetChainID(chainID)
	//获得第一个账户的私钥
	privatekey := utils.GetENV("PRIVATE_KEY_1")
	fmt.Println(privatekey)

	if err != nil {
		panic(err)
	}
	// 构建新交易(gasLimit--gas限制，value--交易值)
	value := big.NewInt(0)
	auth1 := utils.New_auth(client, privatekey, value)

	//部署合约（服务器，区块链ID，合约名称，私钥）
	address, tx0 := utils.Deploy(client, chainID, contract_name, auth1)
	//获取部署合约的gas值
	receipt, err := bind.WaitMined(context.Background(), client, tx0)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("Deploy Gas used: %d\n", receipt.GasUsed)

	//构建调用合约实体
	Contract, err := contract.NewContract(common.HexToAddress(address.Hex()), client)
	if err != nil {
		fmt.Println("实体构建错误: ", err)
	}
	// 构建新交易
	value = big.NewInt(0)
	auth2 := utils.New_auth(client, privatekey, value)

	// 调用set方法设置值
	setValue := "dadasda"
	//调用set函数
	tx, err := Contract.Set(auth2, setValue)
	if err != nil {
		log.Fatalf("Failed to execute set transaction: %v", err)
	}
	receipt, _ = bind.WaitMined(context.Background(), client, tx)
	fmt.Printf("Set() Gas used: %d\n", receipt.GasUsed)

	// fmt.Printf("Set transaction hash: %s\n", tx.Hash().Hex())

	// 调用get方法获取值
	storedValue, err := Contract.Get(&bind.CallOpts{})
	if err != nil {
		log.Fatalf("Failed to execute get call: %v", err)
	}
	fmt.Printf("Stored value: %s\n", storedValue)
}
