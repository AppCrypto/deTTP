package ElGamal_test

import(
      "PRE/crypto/ElGamal"
      "fmt"
      "testing"
      "crypto/rand"
      bn256 "github.com/ethereum/go-ethereum/crypto/bn256/google"
)

var order=bn256.Order

func TestElGamalFunction(t *testing.T){
      //Generates a key pair of the encryptor 
      sk,pk:=ElGamal.EGSetup()
      //Generate a plaintext
      k,_:=rand.Int(rand.Reader, order)
      K:=make([]*bn256.G1,1)
      K[0]= new(bn256.G1).ScalarBaseMult(k)
      //Encrypt the plaintext to ciphertext
      EK:=ElGamal.EGEncrypt(K,pk,len(K))
      //Decrypt the ciphertext
	  _K:=ElGamal.EGDecrypt(EK,sk,len(K))
	  fmt.Printf("解密后的明文信息：%s\n",_K)
}
