package vss_test

import (
	"dttp/crypto/vss"
	"fmt"
	"math/big"
	"testing"
)

func TestMyFunction(t *testing.T) {
	//定义秘密
	// 创建一个新的big.Int实例
	secret := new(big.Int)

	// 使用字符串初始化大整数(10进制字符串)
	secret.SetString("18565186733591291362307462130219129409737445814581163957621748889988504982598", 10)

	//门限值
	threshold := 3
	//多少份份额
	numShares := 7

	// 生成 shares and commitments
	ss, err := vss.GenerateShares(secret, threshold, numShares)
	if err != nil {
		fmt.Println("Error generating shares:", err)
		return
	}

	// 打印 the shares and commitments
	fmt.Println("Shares:")
	for i, share := range ss.Shares {
		fmt.Printf("Share %d: %s\n", i+1, share)
	}

	fmt.Println("\nCommitments:")
	for i, commitment := range ss.Commitments {
		fmt.Printf("Commitment %d: %s\n", i+1, commitment.String())
	}

	// 验证份额
	result:=vss.VerifyShare(ss.Gs, ss.Commitments)
	fmt.Printf("VSS验证结果为：%v\n",result)

	//恢复秘密
	selectedShares := ss.Shares[:threshold]
	selectedIndices := make([]*big.Int, threshold)
	for i := 0; i < threshold; i++ {
		selectedIndices[i] = big.NewInt(int64(i + 1))
	}

	recoveredSecret := vss.RecoverSecret(selectedShares, selectedIndices)
	fmt.Printf("\nRecovered Secret: %s\n", recoveredSecret)
}
