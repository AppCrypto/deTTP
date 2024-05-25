/*
使用说明：

	1、使用前下载solc,abigen配置全局变量，下载命令行ganache
	1、使用时将.sol文件放入compile/contract文件夹，根据.sol文件生成的abi,bin,.go都在此文件夹内。
	2、每个go的实例具体化方法名称都不一样（一般New+合约名），记得查询后修改。

执行命令：

	1、执行./stract.sh————将私钥保存到.env文件内
	2、执行comlile文件夹的main.go————(输入要合约名字)进行compile和abigen
	3、执行ganache --mnemonic "dttp"————开启dttp的gananche
	4、执行main.go
*/
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
	//获得第一个账户的私钥
	privatekey := utils.GetENV("PRIVATE_KEY_1")
	fmt.Println(privatekey)
	privatekey = "75a4b8695e5d35ba5d3f897bb837d03bde178bfe9682ded4687d895e4cf77486"
	if err != nil {
		panic(err)
	}
	// 构建新交易(gasLimit--gas限制，value--交易值)
	value := big.NewInt(0)
	auth1 := utils.New_auth(client, privatekey, chainID, value)

	//部署合约（服务器，区块链ID，合约名称，私钥）
	address := utils.Deploy(client, chainID, contract_name, auth1)

	//构建调用合约实体
	Contract, err := contract.NewContract(common.HexToAddress(address), client)
	if err != nil {
		fmt.Println("实体构建错误: ", err)
	}
	// 构建新交易
	value = big.NewInt(0)
	auth2 := utils.New_auth(client, privatekey, chainID, value)

	// 调用set方法设置值
	setValue := big.NewInt(12345)
	//调用set函数
	tx, err := Contract.Set(auth2, setValue)
	if err != nil {
		log.Fatalf("Failed to execute set transaction: %v", err)
	}
	fmt.Printf("Set transaction hash: %s\n", tx.Hash().Hex())

	// 调用get方法获取值
	storedValue, err := Contract.Get(&bind.CallOpts{})
	if err != nil {
		log.Fatalf("Failed to execute get call: %v", err)
	}
	fmt.Printf("Stored value: %s\n", storedValue.String())
}
