package main

import (
	"context"
	"crypto/ecdsa"
	"dttp/compile/contract"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func getENV(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}
	return os.Getenv(key)
}

func main() {

	client, err := ethclient.Dial(getENV("SERVER"))
	if err != nil {
		panic(err)
	}

	privateKey, err := crypto.HexToECDSA(getENV("PRIVATE_KEY_1"))
	if err != nil {
		panic(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		panic("invalid key")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		panic(err)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println("Gas price:", gasPrice)

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		panic(err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = uint64(0)  // in units
	auth.GasPrice = gasPrice

	address, tx, instance, err := contract.DeployContract(auth, client)
	if err != nil {
		panic(err)
	}

	fmt.Println("Deployed at", address.Hex())

	_, _ = instance, tx

	server := getENV("SERVER")
	fmt.Println("Server: ", server)

	client, err = ethclient.Dial(server)
	if err != nil {
		panic(err)
	}

	// contract := getENV("CONTRACT_ADDRESS")
	conn, err1 := contract.NewContract(common.HexToAddress(address.Hex()), client)
	if err1 != nil {
		panic(err1)
	}
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/greet/:message",
		func(c echo.Context) error {
			message := c.Param("message")
			fmt.Println(message)
			reply, err := conn.Set(&bind.TransactOpts{}, message)
			if err != nil {
				return err
			}
			return c.JSON(http.StatusOK, reply)
		})

	e.GET("/hello", func(c echo.Context) error {
		reply, err := conn.Get(&bind.CallOpts{})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, reply) // Hello World
	})

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
