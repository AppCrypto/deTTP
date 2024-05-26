package ElGamal_test

import(
      "dttp/crypto/ElGmal"
      "fmt"
      "math/big"
      "testing"
)

func TestElGmalFunction(t *testing.T){
      //生成加密者的公私钥对
      sk,pk:=ElGmal.EGSetup()
      //生成一个明文信息
      K：= rand.Int(rand.Reader, order)
      //对该明文信息进行加密
      EK:=EGEncrypt(K,pk)
      //对密文信息进行解密
	_K:=EGDecrypt(EK.EK0,EK.EK1,sk)
}
