package DLEQ

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"strings"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
)

// Q is the order of the integer field (Zq) that fits inside the snark
var Q, _ = new(big.Int).SetString(
	"21888242871839275222246405745257275088696311157297823662689037894645226208583", 10)

// R is the mod of the finite field
var R, _ = new(big.Int).SetString(
	"21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)


func DLEQProofG1(G, H *bn256.G1, xG, xH *bn256.G1, x *big.Int) (c, z *big.Int, rG, rH *bn256.G1, err error) {
	//生成承诺
	r, err := rand.Int(rand.Reader, Q)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	rG = new(bn256.G1).ScalarMult(G, r)
	rH = new(bn256.G1).ScalarMult(H, r)

	// 计算挑战
	new_hash := sha256.New()
	new_hash.Write(xG.Marshal())
	new_hash.Write(xH.Marshal())
	new_hash.Write(rG.Marshal())
	new_hash.Write(rH.Marshal())

	cb := new_hash.Sum(nil)
	c = new(big.Int).SetBytes(cb)
	c.Mod(c, bn256.Order)

	// 生成相应
	z = new(big.Int).Mul(c, x)
	z.Sub(r, z)
	z.Mod(z, R)

	return c, z, rG, rH, nil
}

func DLEQProofG1_G2(G *bn256.G1, H *bn256.G2, xG *bn256.G1, xH *bn256.G2, x *big.Int) (c, z *big.Int, rG *bn256.G1, rH *bn256.G2, err error) {
	//生成承诺
	r, err := rand.Int(rand.Reader, Q)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	rG = new(bn256.G1).ScalarMult(G, r)
	rH = new(bn256.G2).ScalarMult(H, r)

	// 计算挑战
	new_hash := sha256.New()
	new_hash.Write(xG.Marshal())
	new_hash.Write(xH.Marshal())
	new_hash.Write(rG.Marshal())
	new_hash.Write(rH.Marshal())

	cb := new_hash.Sum(nil)
	c = new(big.Int).SetBytes(cb)
	c.Mod(c, R)

	// 生成相应
	z = new(big.Int).Mul(c, x)
	z.Sub(r, z)
	z.Mod(z, R)

	return c, z, rG, rH, nil
}

// Verify verifies the DLEQ proof
func VerifyG1(c, z *big.Int, G, H, xG, xH, rG, rH *bn256.G1) error {
	zG := new(bn256.G1).ScalarMult(G, z)
	zH := new(bn256.G1).ScalarMult(H, z)
	cxG := new(bn256.G1).ScalarMult(xG, c)
	cxH := new(bn256.G1).ScalarMult(xH, c)
	a := new(bn256.G1).Add(zG, cxG)
	b := new(bn256.G1).Add(zH, cxH)
	if !(rG.String() == a.String() && rH.String() == b.String()) {
		return errors.New("invalid proof")
	}
	return nil
}

func VerifyG1_G2(c, z *big.Int, G *bn256.G1, H *bn256.G2, xG *bn256.G1, xH *bn256.G2, rG *bn256.G1, rH *bn256.G2) error {
	zG := new(bn256.G1).ScalarMult(G, z)
	zH := new(bn256.G2).ScalarMult(H, z)
	cxG := new(bn256.G1).ScalarMult(xG, c)
	cxH := new(bn256.G2).ScalarMult(xH, c)
	a := new(bn256.G1).Add(zG, cxG)
	b := new(bn256.G2).Add(zH, cxH)
	if !(rG.String() == a.String() && rH.String() == b.String()) {
		return errors.New("invalid proof")
	}
	return nil
}

func Mul_NewDLEQProof(G, H, xG, xH []*bn256.G1, x []*big.Int) (C, Z []*big.Int, XG, XH, RG, RH []*bn256.G1, err error) {
	k := len(G)
	C = make([]*big.Int, k)
	Z = make([]*big.Int, k)
	XG = make([]*bn256.G1, k)
	XH = make([]*bn256.G1, k)
	RG = make([]*bn256.G1, k)
	RH = make([]*bn256.G1, k)
	var errors []string

	for i := 0; i < k; i++ {
		c, z, rg, rh, err := DLEQProofG1(G[i], H[i], xG[i], xH[i], x[i])
		if err != nil {
			errorMsg := fmt.Sprintf("第%d个proof生成错误: %v", i, err)
			errors = append(errors, errorMsg)
			continue // Optionally skip this index and continue or you can store placeholders
		}
		C[i], Z[i], XG[i], XH[i], RG[i], RH[i] = c, z, xG[i], xH[i], rg, rh
	}

	if len(errors) > 0 {
		return nil, nil, nil, nil, nil, nil, fmt.Errorf("证明生成失败:\n%s", strings.Join(errors, "\n"))
	}
	return C, Z, XG, XH, RG, RH, nil
}

func Mul_Verify(C, Z []*big.Int, G, H, XG, XH, RG, RH []*bn256.G1) error {
	k := len(C)
	var errors []string

	for i := 0; i < k; i++ {
		err := VerifyG1(C[i], Z[i], G[i], H[i], XG[i], XH[i], RG[i], RH[i])
		if err != nil {
			errorMsg := fmt.Sprintf("第%d个proof有问题: %v", i, err)
			errors = append(errors, errorMsg)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("verification failed:\n%s", strings.Join(errors, "\n"))
	}
	return nil
}
