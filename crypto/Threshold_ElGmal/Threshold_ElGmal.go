package Threshold_ElGmal

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/google"
)

var order=bn256.Order

type CK struct {
	CK0 *bn256.G1 
	CK1 *bn256.G1
}

func THEGSetup()(*big.Int, *bn256.G1){
	//生成加密者的公私钥对
	sk,pk,_:=bn256.RandomG1(rand.Reader)
    return sk,pk
}

func THEGEncrypt(m *big.Int, PK *bn256.G1)(*CK){
	fmt.Printf("明文信息为：%s\n",m)
	fmt.Printf("明文映射后的信息为：%s\n",new(bn256.G1).ScalarBaseMult(m).String())
	r,_ := rand.Int(rand.Reader, order)
	ck0:=new(bn256.G1).ScalarBaseMult(r)
	ck1:=new(bn256.G1).Add(new(bn256.G1).ScalarBaseMult(m),new(bn256.G1).ScalarMult(PK,r))
	
	return &CK{
		CK0:ck0,
		CK1:ck1,
	}
}

func THEGKenGen(CK *CK, SK *big.Int, n, t int)([]*bn256.G1){
	VSS_SK,_:=GenerateShares(SK, t, n)
	K := make([]*bn256.G1, n)
	for i:=0;i<n;i++{
		K[i]=new(bn256.G1).ScalarMult(CK.CK0,VSS_SK.Shares[i])
	}
	return K
}



// lagrangeInterpolation 使用拉格朗日插值法恢复密钥的计算
func recoverKey(Key []*bn256.G1, indices []*big.Int, order *big.Int, threshold int)*bn256.G1{
	// k是分享的数量
	k := threshold

	Recover_Key:=new(bn256.G1).ScalarBaseMult(big.NewInt(0))

	// 对于每个分享
	for i := 0; i < k; i++ {
	// 初始化分子（num）和分母（den）为1
	num := big.NewInt(1)
	den := big.NewInt(1)

	// 计算拉格朗日基函数的分子和分母
		for j := 0; j < k; j++ {
			if i != j {
				// 分子累乘 -indices[j]
				num.Mul(num, new(big.Int).Neg(indices[j]))
				num.Mod(num, order)

				// 分母累乘 indices[i] - indices[j]
				den.Mul(den, new(big.Int).Sub(indices[i], indices[j]))
				den.Mod(den, order)
				}
			}
			// 计算分母的逆元（模order）
			den.ModInverse(den, order)
			// 计算每一项的值 shares[i] * num * den
			term := new(big.Int).Mul(big.NewInt(1), num)
			term.Mul(term, den)
			term.Mod(term, order)
			Recover_Key= new(bn256.G1).Add(Recover_Key,new(bn256.G1).ScalarMult(Key[i],term))
	}
	return Recover_Key
}


		

func THEGDecrypt(CK *CK, Key []*bn256.G1, indices []*big.Int, threshold int)(*bn256.G1){
	
	Recover_Key:=recoverKey(Key, indices, order, threshold)
	//解密密文信息
	_m:=new(bn256.G1).Add(CK.CK1, new(bn256.G1).Neg(Recover_Key))
	return _m
}

