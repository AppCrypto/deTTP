package ElGmal

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/google"
)

var order=bn256.Order

type EK struct {
	EK0 *bn256.G1 
	EK1 *bn256.G1
}

func EGSetup()(*big.Int, *bn256.G1){
	//生成加密者的公私钥对
	sk,pk,_:=bn256.RandomG1(rand.Reader)
    return sk,pk
}


func EGEncrypt(K *big.Int, PK *bn256.G1)(*EK){
	fmt.Printf("明文信息为：%s\n",K)
	fmt.Printf("明文映射后的信息为：%s\n",new(bn256.G1).ScalarBaseMult(K).String())
	l,_ := rand.Int(rand.Reader, order)
	ek0:=new(bn256.G1).ScalarBaseMult(l)
	ek1:=new(bn256.G1).Add(new(bn256.G1).ScalarBaseMult(K),new(bn256.G1).ScalarMult(PK,l))
	
	return &EK{
		EK0:ek0,
		EK1:ek1,
	}
}


func EGDecrypt(EK *EK, sk *big.Int)(*bn256.G1){
	//解密密文信息
	_K:=new(bn256.G1).Add(EK.EK1,new(bn256.G1).Neg(new(bn256.G1).ScalarMult(EK.EK0,sk)))
	return _K
}

