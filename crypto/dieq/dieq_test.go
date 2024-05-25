package dieq_test

import (
	"crypto/rand"
	"dttp/crypto/dieq"
	"fmt"
	"math/big"
	"testing"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/google"
)

func TestMyFunction(t *testing.T) {
	//生成随机点g
	g, err := rand.Int(rand.Reader, bn256.Order)
	if err != nil {
		fmt.Println("Failed to generate random H:", err)
		return
	}
	G := new(bn256.G1).ScalarBaseMult(g)
	//生成随机点H
	h, err := rand.Int(rand.Reader, bn256.Order)
	if err != nil {
		fmt.Println("Failed to generate random H:", err)
		return
	}
	H := new(bn256.G1).ScalarBaseMult(h)
	// 创建一个新的big.Int实例，使用字符串初始化大整数(10进制字符串)
	x := new(big.Int)
	x.SetString("18565186733591291362307462130219129409737445814581163957621748889988504982598", 10)
	//生成证明（xH和xG拥有相同的指数x，xH=x*H,xG=x*G）
	c, z, xG, xH, rG, rH, err := dieq.NewDLEQProof(G, H, x)
	if err != nil {
		fmt.Println("Failed to create DLEQ proof:", err)
		return
	}

	rtn := dieq.Verify(c, z, G, H, xG, xH, rG, rH)

	if rtn == nil {
		fmt.Printf("\n\nPeggy has proven she still knows her secret")
	} else {
		fmt.Printf("\n\nProof verification failed: %s\n", rtn)
	}
}
