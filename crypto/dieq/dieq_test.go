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
	}

	// 指定多个实例
	mul_numInstances := 3
	mul_G := make([]*bn256.G1, mul_numInstances)
	mul_H := make([]*bn256.G1, mul_numInstances)
	mul_X := make([]*big.Int, mul_numInstances)

	// 为每个实例生成随机 G, H 和 x
	for i := 0; i < mul_numInstances; i++ {
		g, _ := rand.Int(rand.Reader, bn256.Order)
		mul_G[i] = new(bn256.G1).ScalarBaseMult(g)
		h, _ := rand.Int(rand.Reader, bn256.Order)
		mul_H[i] = new(bn256.G1).ScalarBaseMult(h)
		mul_X[i], _ = rand.Int(rand.Reader, bn256.Order)
	}

	// 生成多个 DLEQ 证明
	mul_C, mul_Z, mul_XG, mul_XH, mul_RG, mul_RH, err := dieq.Mul_NewDLEQProof(mul_G, mul_H, mul_X)
	if err != nil {
		t.Errorf("Failed to create multiple DLEQ proofs: %v", err)
		return
	}

	// 验证生成的证明
	err = dieq.Mul_Verify(mul_C, mul_Z, mul_G, mul_H, mul_XG, mul_XH, mul_RG, mul_RH)
	if err != nil {
		t.Errorf("Verification failed: %v", err)
		return
	}
	// 输出结果
	fmt.Println("All proofs verified successfully.")
}
