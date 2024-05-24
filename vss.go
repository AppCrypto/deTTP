package main

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/google"
)

// 需要运算的参数
type SecretSharing struct {
	Shares      []*big.Int  //密钥分享
	Commitments []*bn256.G1 //承诺
	Gs          []*bn256.G1
	Sgs         []*big.Int
}

// GenerateShares 生成密钥分享和承诺
func GenerateShares(secret *big.Int, threshold, numShares int) (*SecretSharing, error) {
	// 定义曲线的阶数
	order := bn256.Order

	// 生成多项式的随机系数
	coefficients := make([]*big.Int, threshold)
	coefficients[0] = secret
	for i := 1; i < threshold; i++ {
		coefficients[i], _ = rand.Int(rand.Reader, order)
	}

	// 生成密钥分享
	shares := make([]*big.Int, numShares)
	for i := 0; i < numShares; i++ {
		x := big.NewInt(int64(i + 1))
		shares[i] = evaluatePolynomial(coefficients, x, order)
	}

	// 生成承诺
	commitments := make([]*bn256.G1, threshold)
	for i := 0; i < threshold; i++ {
		commitments[i] = new(bn256.G1).ScalarBaseMult(coefficients[i])
	}
	// 生成gs
	gs := make([]*bn256.G1, numShares)
	for i := 0; i < numShares; i++ {
		gs[i] = new(bn256.G1).ScalarBaseMult(shares[i])
	}

	// 生成sgs
	sgs := make([]*big.Int, numShares)
	for i := 0; i < numShares; i++ {
		sgs[i] = new(big.Int).Mul(shares[i], HashBigInt(shares[i]))
	}

	return &SecretSharing{
		Shares:      shares,
		Commitments: commitments,
		Gs:          gs,
		Sgs:         sgs,
	}, nil
}

// hash函数
func HashBigInt(num *big.Int) *big.Int {
	// Convert the big.Int to a byte slice
	numBytes := num.Bytes()

	// Compute the SHA-256 hash
	hash := sha256.Sum256(numBytes)

	// Convert the hash (which is a byte array) back to a big.Int
	hashInt := new(big.Int).SetBytes(hash[:])

	return hashInt
}

// evaluatePolynomial 在给定的 x 处计算多项式的值
func evaluatePolynomial(coefficients []*big.Int, x, order *big.Int) *big.Int {
	result := new(big.Int).Set(coefficients[0])
	xPower := new(big.Int).Set(x)

	for i := 1; i < len(coefficients); i++ {
		term := new(big.Int).Mul(coefficients[i], xPower)
		term.Mod(term, order)
		result.Add(result, term)
		result.Mod(result, order)
		xPower.Mul(xPower, x)
		xPower.Mod(xPower, order)
	}

	return result
}

// VerifyShare 根据承诺验证给定的密钥分享
func VerifyShare(share, x *big.Int, commitments []*bn256.G1) bool {
	left := new(bn256.G1).ScalarBaseMult(share)
	right := new(bn256.G1)

	xPower := big.NewInt(1)
	for _, commitment := range commitments {
		temp := new(bn256.G1).ScalarMult(commitment, xPower)
		right.Add(right, temp)
		xPower.Mul(xPower, x)
		xPower.Mod(xPower, bn256.Order)
	}

	return left.String() == right.String()
}

// lagrangeInterpolation 使用拉格朗日插值法恢复密钥
func lagrangeInterpolation(shares []*big.Int, indices []*big.Int, order *big.Int) *big.Int {
	secret := big.NewInt(0)
	k := len(shares)

	for i := 0; i < k; i++ {
		num := big.NewInt(1)
		den := big.NewInt(1)

		for j := 0; j < k; j++ {
			if i != j {
				num.Mul(num, new(big.Int).Neg(indices[j]))
				num.Mod(num, order)

				den.Mul(den, new(big.Int).Sub(indices[i], indices[j]))
				den.Mod(den, order)
			}
		}

		den.ModInverse(den, order)
		term := new(big.Int).Mul(shares[i], num)
		term.Mul(term, den)
		term.Mod(term, order)
		secret.Add(secret, term)
		secret.Mod(secret, order)
	}

	return secret
}

// RecoverSecret 使用拉格朗日插值法从密钥分享中恢复密钥
func RecoverSecret(shares []*big.Int, indices []*big.Int) *big.Int {
	order := bn256.Order
	return lagrangeInterpolation(shares, indices, order)
}

func main() {
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
	ss, err := GenerateShares(secret, threshold, numShares)
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
	fmt.Println("\nVerifying shares:")
	for i, share := range ss.Shares {
		x := big.NewInt(int64(i + 1))
		if VerifyShare(share, x, ss.Commitments) {
			fmt.Printf("Share %d is valid\n", i+1)
		} else {
			fmt.Printf("Share %d is invalid\n", i+1)
		}
	}

	//恢复秘密
	selectedShares := ss.Shares[:threshold]
	selectedIndices := make([]*big.Int, threshold)
	for i := 0; i < threshold; i++ {
		selectedIndices[i] = big.NewInt(int64(i + 1))
	}

	recoveredSecret := RecoverSecret(selectedShares, selectedIndices)
	fmt.Printf("\nRecovered Secret: %s\n", recoveredSecret)

}
