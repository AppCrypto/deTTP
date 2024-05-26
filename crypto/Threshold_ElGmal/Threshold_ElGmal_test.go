package Threshold_ElGmal_test

import(
  "dttp/crypto/Threshold_ElGmal"
  "fmt"
  "math/big"
  "testing"
)

func TestThresholdElGmalFunction(t *testing.T){
      //定义加密者的公私钥
    	sko,pko:=THEGSetup()
      //随机生成一个明文信息
	    m,_ := rand.Int(rand.Reader, order)
      //加密明文信息
	    CK:=THEGEncrypt(m,pko)
      //生成加密的密钥份额
	    Key:=THEGKenGen(CK, sko, numShares, threshold)
      //解密密文信息
	    _m:=THEGDecrypt(CK, Key, selectedIndices, threshold)
	    fmt.Printf("解密信息为：%s\n",_m.String())
}
