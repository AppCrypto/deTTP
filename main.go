package main

import (
	"fmt"
	"PRE/crypto/KZG"
	"PRE/crypto/ElGamal"
	"math/big"
	"crypto/rand"
	"PRE/crypto/DLEQ"
	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
	/*
	"context"
	"PRE/compile/contract"
	"PRE/utils"
	"log"
	
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	//"github.com/ethereum/go-ethereum/core/types"
	*/
)

// Q is the order of the integer field (Zq) that fits inside the snark
var Q, _ = new(big.Int).SetString(
	"21888242871839275222246405745257275088696311157297823662689037894645226208583", 10)

// R is the mod of the finite field
var R, _ = new(big.Int).SetString(
	"21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)

type RK struct {

	RK0 []*bn256.G1
	RK1 []*bn256.G1
}

type DLEQProofG1_G2 struct {
	C  *big.Int
	Z  *big.Int
	RG *bn256.G1
	RH *bn256.G2
}
/*
var order = bn256.Order

type DLEQProofs struct {
	C  []*big.Int
	Z  []*big.Int
	XG []*bn256.G1
	XH []*bn256.G1
	RG []*bn256.G1
	RH []*bn256.G1
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
*/

func main() {
//==========================================System Initialization===========================================//
	// the number of key shares
	numShares := 10
	// threshold value
	threshold := numShares/2 + 1
	//threshold := 2*numShares/3 + 1

	fmt.Printf("The number of shares is %v\n", numShares)
	fmt.Printf("The threshold value is %v\n", threshold)

	//Setup Algorithm
	PK,_:= KZG.NewTrustedSetup(numShares)
	//fmt.Printf("The system public key is %v\n",PK)

//===========================================User Registration===========================================//
	//KeyGen_u
	//The key pair of data owner
	sko,pko:=ElGamal.EGKeyGen()
	vko:=new(bn256.G2).ScalarBaseMult(sko)	
	fmt.Printf("The key pair of data owner is %v || %v || %v\n", sko,pko,vko)
	//DLEQProof(g1,pko,g2,vko,sko)
	var prf_sko DLEQProofG1_G2
	c,z,rg,rh,_:= DLEQ.DLEQProofG1_G2(PK.Tau1[0], PK.Tau2[0], pko, vko, sko)
	prf_sko.C=c
	prf_sko.Z=z
	prf_sko.RG=rg
	prf_sko.RH=rh
	fmt.Printf("The DLEQ proof of vko is %v\n",prf_sko)
	//KeyGen_u
	//The key pair of data user
	sku,pku:=ElGamal.EGKeyGen()
	vku:=new(bn256.G2).ScalarBaseMult(sku)
	fmt.Printf("The key pair of data user is %v || %v || %v\n", sku,pku,vku)
	var prf_sku DLEQProofG1_G2
	c,z,rg,rh,_= DLEQ.DLEQProofG1_G2(PK.Tau1[0], PK.Tau2[0], pku, vku, sku)
	prf_sku.C=c
	prf_sku.Z=z
	prf_sku.RG=rg
	prf_sku.RH=rh
	fmt.Printf("The DLEQ proof of vku is %v\n",prf_sku)
	//KeyGen_DR	
	//The key pair of data regulators
	sk:=make([]*big.Int,numShares)
	pk:=make([]*bn256.G1,numShares)
	for i:=0;i<numShares;i++{
		sk[i],pk[i]=ElGamal.EGKeyGen()
		fmt.Printf("The key pair of data regulator%v is %v || %v\n", i,sk[i],pk[i])
	}
	
	//VKVerify
	//vko verify
	DLEQResult := DLEQ.VerifyG1_G2(prf_sko.C, prf_sko.Z, PK.Tau1[0], PK.Tau2[0], pko, vko, prf_sko.RG, prf_sko.RH)

	if DLEQResult == nil {
		fmt.Printf("\n\nThe verification of vko passes!\n\n")
	}else{
		fmt.Printf("\n\nThe verification of vko fails to pass!\n\n")
	}
	//vku verify
	DLEQResult = DLEQ.VerifyG1_G2(prf_sku.C, prf_sku.Z, PK.Tau1[0], PK.Tau2[0], pku, vku, prf_sku.RG, prf_sku.RH)

	if DLEQResult == nil {
		fmt.Printf("\n\nThe verification of vku passes!\n\n")
	}else{
		fmt.Printf("\n\nThe verification of vku fails to pass!\n\n")
	}
//======================================Sensitive Message Encryption========================================//
	//Data owner encrypts the sensitive message M which is the AES key.
	//secret=H()
	secret:=big.NewInt(int64(2))
	m,_ := rand.Int(rand.Reader, Q)
	//gT=e(g1,g2)
	M:=bn256.Pair(new(bn256.G1).ScalarBaseMult(m),new(bn256.G2).ScalarBaseMult(big.NewInt(int64(1))))
	fmt.Printf("The plaintext is %v\n",M)
	C:= ElGamal.EGEncrypt(M,pko,secret)
	fmt.Printf("The data ciphertext is %v\n",C)
	/*
	//Pair Check
	K:=new(bn256.G1).ScalarMult(pko,s)
	result:=ElGamal.EGDecrypt(C,K)
	fmt.Printf("The decryption result is %v\n",result)
	*/
//======================================Re-encrypted Key Generation========================================//

	//Generate the polynomial coefficients
	coefficients := make([]*big.Int, threshold)
	coefficients[0] = secret
	for i := 1; i < threshold; i++ {
		coefficients[i], _ = rand.Int(rand.Reader, Q)
	}

	//Compute the shares of secret
	shares := make([]*big.Int, numShares)
	for i := 0; i < numShares; i++ {
		x := big.NewInt(int64(i + 1))
		shares[i] = KZG.EvaluatePolynomial(coefficients, x)
	}
	fmt.Printf("The shares is %v\n",shares)	
	
	//Generate RK
	var ReKey RK
	ReKey0:=make([]*bn256.G1,numShares)
	ReKey1:=make([]*bn256.G1,numShares)
	Base:=make([]*bn256.G1,numShares)
	for i:=0;i<numShares;i++{
		ReKey0[i]=new(bn256.G1).ScalarBaseMult(shares[i])
		Base[i]=new(bn256.G1).Add(pko,new(bn256.G1).Add(pk[i],pku))
		ReKey1[i]=new(bn256.G1).ScalarMult(Base[i],shares[i])
	}
	ReKey.RK0=ReKey0
	ReKey.RK1=ReKey1
	fmt.Printf("The re-encrypted key shares are %v\n",ReKey)
	
	//Generate Proof
	//KZG commitment
	Commit:=KZG.Commit(PK,coefficients)
	fmt.Printf("The commitment is %v\n",Commit)
	//KZG witness
	witness:=make([]*bn256.G1,numShares)
	for i:=0;i<numShares;i++{
		x := big.NewInt(int64(i + 1))
		witness[i],_=KZG.EvaluationProof(PK, coefficients, x, shares[i])
	}
	fmt.Printf("The witness is %v\n",witness)
	
}
