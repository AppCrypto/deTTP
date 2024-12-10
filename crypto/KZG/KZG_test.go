package KZG_test

import(
	"PRE/crypto/KZG"
	"fmt"
	"testing"
	"math/big"
	//"crypto/rand"
	//bn256 "github.com/ethereum/go-ethereum/crypto/bn256"
)

//var order=bn256.Order

func TestKZGFunction(t *testing.T){
	PK,_:= KZG.NewTrustedSetup(4)
	fmt.Printf("The system public key(G1) is: %s\n",PK.Tau1)
	fmt.Printf("The system public key(G2) is: %s\n",PK.Tau2)
	// p(x) = x^3 + x + 5
	p := []*big.Int{
		big.NewInt(5),
		big.NewInt(1), // x^1
		big.NewInt(0), // x^2
		big.NewInt(1), // x^3
	}
	//Commitment Generation
	C:=KZG.Commit(PK,p)
	fmt.Printf("The commitment is: %s\n",C)
	//Witness Generation
	//Lack the computation:p(x)-->y
	x := big.NewInt(3)
	fmt.Printf("The cofficients is %s\n",p)
	y:=KZG.EvaluatePolynomial(p,x)
	fmt.Printf("The value of y is: %s\n",y)
	w,_:=KZG.EvaluationProof(PK, p, x, y)
	fmt.Printf("The witness is: %s\n",w)
	//Verify
	result:=KZG.Verify(PK, C, w, x, y)
	fmt.Printf("The verification result is: %v\n",result)
}
