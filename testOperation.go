package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/google"
)

var order = bn256.Order

func main() {

	l, _ := rand.Int(rand.Reader, order)
	var n int64 = 1000
	starttime := time.Now().UnixMicro()
	for i := 0; i < int(n); i++ {
		new(bn256.G1).ScalarBaseMult(l)
	}
	endtime := time.Now().UnixMicro()
	fmt.Printf("exponentiation time cost %d us\n", (endtime-starttime)/n)

	g1l := new(bn256.G1).ScalarBaseMult(l)
	r, _ := rand.Int(rand.Reader, bn256.Order)
	g2r := new(bn256.G2).ScalarBaseMult(r)
	starttime = time.Now().UnixMicro()
	for i := 0; i < int(n); i++ {
		bn256.Pair(g1l, g2r)
	}
	endtime = time.Now().UnixMicro()
	fmt.Printf("pairing time cost %d us\n", (endtime-starttime)/n)

	x, _ := rand.Int(rand.Reader, order)
	y, _ := rand.Int(rand.Reader, order)
	num := big.NewInt(1)
	starttime = time.Now().UnixNano()
	for i := 0; i < int(n); i++ {
		new(big.Int).Neg(x)
	}
	endtime = time.Now().UnixNano()
	fmt.Printf("Neg time cost %d ns\n", (endtime-starttime)/n)

	starttime = time.Now().UnixNano()
	for i := 0; i < int(n); i++ {
		num.Mul(num, x)
	}
	endtime = time.Now().UnixNano()
	fmt.Printf("Mul time cost %d ns\n", (endtime-starttime)/n)

	starttime = time.Now().UnixNano()
	for i := 0; i < int(n); i++ {
		num.Mod(y, order)
	}
	endtime = time.Now().UnixNano()
	fmt.Printf("Mod time cost %d ns\n", (endtime-starttime)/n)
}
