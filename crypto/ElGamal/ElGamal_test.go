package ElGamal_test

import(
      "dttp/crypto/ElGamal"
      "fmt"
      "testing"
      "crypto/rand"
      bn256 "github.com/ethereum/go-ethereum/crypto/bn256/google"
)

var order=bn256.Order

func TestElGamalFunction(t *testing.T){
      //生成加密者的公私钥对
      sk,pk:=ElGamal.EGSetup()
      //生成一个明文信息
      k,_:=rand.Int(rand.Reader, order)
      K:=make([]*bn256.G1,1)
      K[0]= new(bn256.G1).ScalarBaseMult(k)
      //对该明文信息进行加密
      EK:=ElGamal.EGEncrypt(K,pk,len(K))
      //对密文信息进行解密
	  _K:=ElGamal.EGDecrypt(EK,sk,len(K))
	  fmt.Printf("解密后的明文信息：%s\n",_K)
}
