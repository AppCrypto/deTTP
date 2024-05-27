package Threshold_ElGamal_test

import(
  "dttp/crypto/Threshold_ElGamal"
  "fmt"
  "math/big"
  "crypto/rand"
  "testing"
  bn256 "github.com/ethereum/go-ethereum/crypto/bn256/google"
)

var order=bn256.Order

func TestThresholdElGmalFunction(t *testing.T){
     //门限值
     threshold := 3
     //多少份份额
     numShares := 7
	
     selectedIndices := make([]*big.Int, threshold)
     for i := 0; i < threshold; i++ {
	selectedIndices[i] = big.NewInt(int64(i + 1))
     }
      //定义加密者的公私钥
        sko,pko:=Threshold_ElGamal.THEGSetup()
      //随机生成一个明文信息
	m,_ := rand.Int(rand.Reader, order)
      //加密明文信息
	C:=Threshold_ElGamal.THEGEncrypt(m,pko)
      //生成加密的密钥份额
	VSS_SK,Key:=Threshold_ElGamal.THEGKenGen(C, sko, numShares, threshold)
	fmt.Printf("映射后的秘密份额为：%s\n", VSS_SK.Gs)
     //解密密文信息
	_m:=Threshold_ElGamal.THEGDecrypt(C, Key, selectedIndices, threshold)
	fmt.Printf("解密后的信息为：%s\n",_m.String())
	
}


