package main

import (
	"dttp/contract"
	"dttp/utils"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

/*
	使用前下载solc。  abigen的exe文件已经放在contract文件夹内
	1、使用时将.sol文件放入contract文件夹，根据.sol文件生成的abi,bin,.go都在此文件夹内。
	2、需要修改区块链服务器的地址，ID等
	如果是ganache：
	chainID := big.NewInt(1337),自动获取可能会报错。
	自动获取chainID:
	chainID, err := client.NetworkID(context.Background())
    if err != nil {
        log.Fatalf("Failed to get network ID: %v", err)
    }
	3、每个go的实例具体化方法名称都不一样（一般New+合约名），记得查询后修改。
*/

func main() {
	// 连接到Ganache
	client, err := ethclient.Dial("http://127.0.0.1:7545")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	//区块链ID-------------------------------------------------(修改处)
	chainID := big.NewInt(1337)

	//合约名称-------------------------------------------------(修改处)

	contract_name := "SimpleStorage"

	//编译合约（合约名称）
	utils.Compile(contract_name)

	//私钥-------------------------------------------------(修改处)
	privatekey := "096bfb23f5bf0a2486d14370eaf0b29b93dcff7c6e35c07bf4d49f95e8b15fe0"

	// 构建新交易(gasLimit--gas限制，value--交易值)
	gasLimit := uint64(3000000)
	value := big.NewInt(0)
	auth1 := utils.New_auth(client, privatekey, chainID, gasLimit, value)

	//部署合约（服务器，区块链ID，合约名称，私钥）
	address := utils.Deploy(client, chainID, contract_name, auth1)

	//将合约变为go语言接口
	utils.Abigen(contract_name)

	//构建调用合约实体
	Contract, err := contract.NewContract(common.HexToAddress(address), client)
	if err != nil {
		fmt.Println("实体构建错误: ", err)
	}

	// 构建新交易
	gasLimit = uint64(3000000)
	value = big.NewInt(0)
	auth2 := utils.New_auth(client, privatekey, chainID, gasLimit, value)

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
