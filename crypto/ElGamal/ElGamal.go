package ElGamal

import (
	"crypto/rand"
	//"fmt"
	"math/big"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/google"
)

var order=bn256.Order

type EK struct {
	EK0 []*bn256.G1 
	EK1 []*bn256.G1
}

func EGSetup()(*big.Int, *bn256.G1){
	//Generate a key pair
	sk,pk,_:=bn256.RandomG1(rand.Reader)
    return sk,pk
}

func EGEncrypt(K []*bn256.G1, PK *bn256.G1, numShares int)(*EK){
	ek0:=make([]*bn256.G1,numShares)
	ek1:=make([]*bn256.G1,numShares)
	//fmt.Printf("The plaintext is %s\n",K)
	l,_ := rand.Int(rand.Reader, order)
	for i:=0;i<numShares;i++{
		ek0[i]=new(bn256.G1).ScalarBaseMult(l)
		ek1[i]=new(bn256.G1).Add(K[i],new(bn256.G1).ScalarMult(PK,l))
	}
	return &EK{
		EK0:ek0,
		EK1:ek1,
	}
}


func EGDecrypt(EK *EK, sk *big.Int, numShares int)([]*bn256.G1){
	_K:=make([]*bn256.G1,numShares)
	for i:=0;i<numShares;i++{
		_K[i]=new(bn256.G1).Add(EK.EK1[i],new(bn256.G1).Neg(new(bn256.G1).ScalarMult(EK.EK0[i],sk)))
	}
	return _K
}
