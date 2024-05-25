package dieq

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"math/big"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/google"
)

func NewDLEQProof(G, H *bn256.G1, x *big.Int) (c, z *big.Int, xG, xH, rG, rH *bn256.G1, err error) {
	//加密x
	xG = new(bn256.G1).ScalarMult(G, x)
	xH = new(bn256.G1).ScalarMult(H, x)
	//生成承诺
	r, err := rand.Int(rand.Reader, bn256.Order)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
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
	z.Mod(z, bn256.Order)

	return c, z, xG, xH, rG, rH, nil
}

// Verify verifies the DLEQ proof
func Verify(c, z *big.Int, G, H, xG, xH, rG, rH *bn256.G1) error {
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
