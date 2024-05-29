// go与区块链交互需要的函数
package utils

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

var chainID *big.Int

// 将contract文件夹下的合约部署在区块链上
func Deploy(client *ethclient.Client, chainID *big.Int, contract_name string, auth *bind.TransactOpts) (common.Address, *types.Transaction) {
	// 读取智能合约的 ABI 和字节码
	abiBytes, err := os.ReadFile("compile/contract/" + contract_name + ".abi")
	if err != nil {
		log.Fatalf("Failed to read ABI file: %v", err)
	}

	bin, err := os.ReadFile("compile/contract/" + contract_name + ".bin")
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
	return address, tx
}

// 创建交易签名
func New_auth(client *ethclient.Client, privatekey string, value *big.Int) *bind.TransactOpts {
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
	gasLimit := uint64(3000000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("Failed to get gas price: %v", err)
	}

	// 发送交易
	auth, err := bind.NewKeyedTransactorWithChainID(key, GetChainID())
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = value
	auth.GasLimit = gasLimit
	auth.GasPrice = gasPrice
	return auth
}
func GetChainID() *big.Int {
	return chainID
}

func SetChainID(chid *big.Int) {
	chainID = chid
}

// 读取.env文件
func GetENV(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}
	return os.Getenv(key)
}
