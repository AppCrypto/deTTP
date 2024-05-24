// go与区块链交互需要的函数
package go_contract

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/exec"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// 将contract文件夹下的合约编译
func Compile(contract_name string) {
	// 定义要执行的命令和参数
	cmd := exec.Command("solc", "--abi", "--bin", "-o", ".", contract_name+".sol", "--overwrite")
	cmd.Dir = "./contract"
	// 捕获标准输出和标准错误
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	// 执行命令
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to execute command: %v, %s", err, stderr.String())
	}
	// 打印标准输出
	fmt.Println(out.String())
}

// 将contract文件夹下的合约部署在区块链上
func Deploy(client *ethclient.Client, chainID *big.Int, contract_name string, auth *bind.TransactOpts) string {
	// 读取智能合约的 ABI 和字节码
	abiBytes, err := os.ReadFile("contract/" + contract_name + ".abi")
	if err != nil {
		log.Fatalf("Failed to read ABI file: %v", err)
	}

	bin, err := os.ReadFile("contract/" + contract_name + ".bin")
	if err != nil {
		log.Fatalf("Failed to read BIN file: %v", err)
	}

	// 解析 ABI
	parsedABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		log.Fatalf("Failed to parse ABI: %v", err)
	}

	address, tx, _, err := bind.DeployContract(auth, parsedABI, common.FromHex(string(bin)), client)
	if err != nil {
		log.Fatalf("Failed to deploy contract: %v", err)
	}
	fmt.Printf("Contract deployed! Address: %s\n", address.Hex())
	fmt.Printf("Transaction hash: %s\n", tx.Hash().Hex())
	return address.Hex()
}

// 根据contract文件夹下的.sol文件生成.go文件，变成go语言可以访问的接口
func Abigen(contract_name string) {
	// 定义要执行的命令和参数
	cmd := exec.Command("./"+"abigen", "--bin="+contract_name+".bin", "--abi="+contract_name+".abi", "--pkg=contract", "--out="+contract_name+".go")
	cmd.Dir = "./contract"
	// 捕获标准输出和标准错误
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	// 执行命令
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to execute command: %v, %s", err, stderr.String())
	}
	// 打印标准输出
	fmt.Println(out.String())
	fmt.Println("abigen is successs")
}

// 创建交易签名
func New_auth(client *ethclient.Client, privatekey string, chainID *big.Int, gasLimit uint64, value *big.Int) *bind.TransactOpts {
	// 获取账户密钥
	key, err := crypto.HexToECDSA(privatekey)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	// 获取账户地址
	publicKey := key.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatalf("Failed to cast public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 获取 nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("Failed to get nonce: %v", err)
	}

	// 设置交易参数

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("Failed to get gas price: %v", err)
	}

	// 发送交易
	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = value
	auth.GasLimit = gasLimit
	auth.GasPrice = gasPrice
	return auth
}
