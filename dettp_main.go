package main

import (
	"context"
	"crypto/rand"
	"dttp/compile/contract"
	"dttp/crypto/ElGamal"
	"dttp/crypto/ThresholdElGamal"
	"dttp/crypto/dleq"
	"dttp/crypto/vss"
	"dttp/utils"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/google"
	"github.com/ethereum/go-ethereum/ethclient"
	//"github.com/ethereum/go-ethereum/core/types"
)

var order = bn256.Order

type DLEQProofs struct {
	C  []*big.Int
	Z  []*big.Int
	XG []*bn256.G1
	XH []*bn256.G1
	RG []*bn256.G1
	RH []*bn256.G1
}

type DLEQProof struct {
	C  *big.Int
	Z  *big.Int
	XG *bn256.G1
	XH *bn256.G1
	RG *bn256.G1
	RH *bn256.G1
}

type EK struct {
	EK0 []*bn256.G1
	EK1 []*bn256.G1
}

func G1ToG1Point(bn256Point *bn256.G1) contract.VerificationG1Point {
	// Marshal the G1 point to get the X and Y coordinates as bytes
	point := bn256Point.Marshal()

	// Create big.Int for X and Y coordinates
	x := new(big.Int).SetBytes(point[:32])
	y := new(big.Int).SetBytes(point[32:64])

	g1Point := contract.VerificationG1Point{
		X: x,
		Y: y,
	}
	return g1Point
}

func G1ToBigIntArray(point *bn256.G1) [2]*big.Int {
	// Marshal the G1 point to get the X and Y coordinates as bytes
	pointBytes := point.Marshal()

	// Create big.Int for X and Y coordinates
	x := new(big.Int).SetBytes(pointBytes[:32])
	y := new(big.Int).SetBytes(pointBytes[32:64])

	return [2]*big.Int{x, y}
}

func main() {

	// the number of key shares
	numShares := 10
	// threshold value
	threshold := numShares/2 + 1
	//threshold := 2*numShares/3 + 1

	var n int64 = 1

	fmt.Printf("The number of shares is %v\n", numShares)
	fmt.Printf("The threshold value is %v\n", threshold)

	//------------------------------------------Registration-------------------------------------//
	//The public parameters
	g1 := new(bn256.G1)
	g1Scalar := big.NewInt(1)
	g := g1.ScalarBaseMult(g1Scalar)


	//Data owner's key pair (sko,pko) and the public key pko is published on the blockchain
	sko, pko := ThresholdElGamal.THEGSetup()
	auth1 := utils.Transact(client, privatekey, big.NewInt(0))
	tx1, _ := Contract.UploadOwnerPk(auth1, G1ToG1Point(pko))


	//Proxies' key pairs (SKs, PKs) and these public keys PKs are published on the blockchain
	SKs := make([]*big.Int, numShares)  //the set of TTPs' private key
	PKs := make([]*bn256.G1, numShares) //the set of TTPs' public key
	var TTPs_PKs []contract.VerificationG1Point

	for i := 0; i < numShares; i++ {
		sk, pk, _ := bn256.RandomG1(rand.Reader)
		SKs[i] = sk
		PKs[i] = pk

		g1Point = G1ToG1Point(pk)
		TTPs_PKs = append(TTPs_PKs, g1Point)

	}

	//Data user's key pair and the public key is published on the blockchain

	sku, pku := ElGamal.EGSetup()
	auth3 := utils.Transact(client, privatekey, big.NewInt(0))
	tx3, _ := Contract.UploadUserPk(auth3, G1ToG1Point(pku))

	receipt3, err := bind.WaitMined(context.Background(), client, tx3)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("Upload User's pk Gas used: %d\n", receipt3.GasUsed)
	//---------------------------------------Secret-Hiding-----------------------------------------//


}
