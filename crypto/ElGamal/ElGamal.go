package ElGamal

import (
	"crypto/rand"
	//"fmt"
	"math/big"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"

)

//var R, _ = new(big.Int).SetString(
	//"21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)

type C struct {
	C0 *bn256.G2 
	C1 *bn256.GT
}

func EGKeyGen()(*big.Int, *bn256.G1){
	//Generate a key pair
	sk,_ := rand.Int(rand.Reader, Q)
	pk:=new(bn256.G1).ScalarBaseMult(sk)
	//vk:=new(bn256.G2).ScalarBaseMult(sk)
    return sk,pk
}

func EGEncrypt(m *bn256.GT,pk *bn256.G1,s *big.Int)(*C){
	//fmt.Printf("The plaintxt is %s\n",new(bn256.G1).ScalarBaseMult(m).String())
	r, _ := rand.Int(rand.Reader, Q)
	c0 := new(bn256.G2).ScalarBaseMult(r)
	c1 := new(bn256.GT).Add(m, bn256.Pair(new(bn256.G1).ScalarMult(pk,s), c0))

	return &C{
		C0: c0,
		C1: c1,
	}
}


func EGDecrypt(C *C, K *bn256.G1)(*bn256.GT){
	m:=new(bn256.GT).Add(C.C1, new(bn256.GT).Neg(bn256.Pair(K,C.C0)))
	return m
}

