package main

import (
	"context"
	"crypto/rand"
	"dttp/compile/contract"
	"dttp/crypto/ElGamal"
	"dttp/crypto/ThresholdElGamal"
	"dttp/crypto/dleq"
	"dttp/crypto/vss"
	"dttp/utils"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/google"
	"github.com/ethereum/go-ethereum/ethclient"
	//"github.com/ethereum/go-ethereum/core/types"
)

var order = bn256.Order

type DLEQProofs struct {
	C  []*big.Int
	Z  []*big.Int
	XG []*bn256.G1
	XH []*bn256.G1
	RG []*bn256.G1
	RH []*bn256.G1
}

type DLEQProof struct {
	C  *big.Int
	Z  *big.Int
	XG *bn256.G1
	XH *bn256.G1
	RG *bn256.G1
	RH *bn256.G1
}

type EK struct {
	EK0 []*bn256.G1
	EK1 []*bn256.G1
}

func G1ToG1Point(bn256Point *bn256.G1) contract.VerificationG1Point {
	// Marshal the G1 point to get the X and Y coordinates as bytes
	point := bn256Point.Marshal()

	// Create big.Int for X and Y coordinates
	x := new(big.Int).SetBytes(point[:32])
	y := new(big.Int).SetBytes(point[32:64])

	g1Point := contract.VerificationG1Point{
		X: x,
		Y: y,
	}
	return g1Point
}

func G1ToBigIntArray(point *bn256.G1) [2]*big.Int {
	// Marshal the G1 point to get the X and Y coordinates as bytes
	pointBytes := point.Marshal()

	// Create big.Int for X and Y coordinates
	x := new(big.Int).SetBytes(pointBytes[:32])
	y := new(big.Int).SetBytes(pointBytes[32:64])

	return [2]*big.Int{x, y}
}

func main() {

	contract_name := "Verification"
	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	privatekey := utils.GetENV("PRIVATE_KEY_1")

	auth := utils.Transact(client, privatekey, big.NewInt(0))

	address, tx := utils.Deploy(client, contract_name, auth)

	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("Deploy Gas used: %d\n", receipt.GasUsed)

	Contract, err := contract.NewContract(common.HexToAddress(address.Hex()), client)
	if err != nil {
		fmt.Println(err)
	}

	// the number of key shares
	numShares := 10
	// threshold value
	threshold := numShares/2 + 1
	//threshold := 2*numShares/3 + 1

	var n int64 = 1

	fmt.Printf("The number of shares is %v\n", numShares)
	fmt.Printf("The threshold value is %v\n", threshold)
	var g1Point contract.VerificationG1Point

	//------------------------------------------Registration-------------------------------------//
	//The public parameters
	g1 := new(bn256.G1)
	g1Scalar := big.NewInt(1)
	g := g1.ScalarBaseMult(g1Scalar)

	auth0 := utils.Transact(client, privatekey, big.NewInt(0))
	tx0, _ := Contract.UploadGenerator(auth0, G1ToG1Point(g))

	receipt0, err := bind.WaitMined(context.Background(), client, tx0)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("Upload the generator Gas used: %d\n", receipt0.GasUsed)

	//Data owner's key pair (sko,pko) and the public key pko is published on the blockchain
	sko, pko := ThresholdElGamal.THEGSetup()
	auth1 := utils.Transact(client, privatekey, big.NewInt(0))
	tx1, _ := Contract.UploadOwnerPk(auth1, G1ToG1Point(pko))

	receipt1, err := bind.WaitMined(context.Background(), client, tx1)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("Upload Owner's pk Gas used: %d\n", receipt1.GasUsed)

	//TTPs' key pairs (SKs, PKs) and these public keys PKs are published on the blockchain
	SKs := make([]*big.Int, numShares)  //the set of TTPs' private key
	PKs := make([]*bn256.G1, numShares) //the set of TTPs' public key
	var TTPs_PKs []contract.VerificationG1Point

	for i := 0; i < numShares; i++ {
		sk, pk, _ := bn256.RandomG1(rand.Reader)
		SKs[i] = sk
		PKs[i] = pk

		g1Point = G1ToG1Point(pk)
		TTPs_PKs = append(TTPs_PKs, g1Point)

	}

	//TODO(Figure 6)： Test the gas comsuption of uploading TTPs' PKs with the number challenge of TTPs
	auth2 := utils.Transact(client, privatekey, big.NewInt(0))
	tx2, _ := Contract.UploadTTPsPk(auth2, TTPs_PKs)

	receipt2, err := bind.WaitMined(context.Background(), client, tx2)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("Figure 6: Upload TTP 's pk Gas used: %d\n", receipt2.GasUsed)
	//Data user's key pair and the public key is published on the blockchain

	sku, pku := ElGamal.EGSetup()
	auth3 := utils.Transact(client, privatekey, big.NewInt(0))
	tx3, _ := Contract.UploadUserPk(auth3, G1ToG1Point(pku))

	receipt3, err := bind.WaitMined(context.Background(), client, tx3)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("Upload User's pk Gas used: %d\n", receipt3.GasUsed)
	//---------------------------------------Secret-Hiding-----------------------------------------//
	// //Randomly generate a plaintext m
	m, _ := rand.Int(rand.Reader, order)
	//Data owner encrypts the plaintext to ciphertext C and C is published on the blockchain
	C := ThresholdElGamal.THEGEncrypt(m, pko)
	fmt.Printf("The ciphertext C is %s\n", C)

	auth4 := utils.Transact(client, privatekey, big.NewInt(0))
	tx4, _ := Contract.UploadCiphertext(auth4, G1ToG1Point(C.C0), G1ToG1Point(C.C1))
	receipt4, err := bind.WaitMined(context.Background(), client, tx4)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("upload Ciphertext C Gas used: %d\n", receipt4.GasUsed)
	//Generate the PVSS shares Key of sko, the share commitment Commitments and {g^si} and publish Commitments and {g^si} on the blockchain
	//TODO(Figure 4):test the time cost of THEGKenGen algorithm with the change of TTPs' number
	var VSS_SK *vss.SecretSharing
	var Key []*bn256.G1
	starttime := time.Now().UnixMilli()
	for i := 0; i < int(n); i++ {
		VSS_SK, Key = ThresholdElGamal.THEGKenGen(C, sko, numShares, threshold)
	}
	endtime := time.Now().UnixMilli()
	fmt.Printf("Figure 4: the time cost of THEGKenGen is %v ms\n", (endtime-starttime)/n)

	var Gs []contract.VerificationG1Point
	var Commitments []contract.VerificationG1Point
	for i := 0; i < numShares; i++ {
		g1Point = G1ToG1Point(VSS_SK.Gs[i])
		Gs = append(Gs, g1Point)
	}
	auth5 := utils.Transact(client, privatekey, big.NewInt(0))
	tx5, _ := Contract.UploadGs(auth5, Gs)
	receipt5, err := bind.WaitMined(context.Background(), client, tx5)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("upload Gs Gas used: %d\n", receipt5.GasUsed)
	for i := 0; i < threshold; i++ {
		g1Point = G1ToG1Point(VSS_SK.Commitments[i])
		Commitments = append(Commitments, g1Point)
	}
	//Data owner uses the TTPs' public keys to encrypt Key to CKey and publishes CKey on the blockchain
	CKeys := make([]*bn256.G1, numShares)
	for i := 0; i < numShares; i++ {
		CKeys[i] = new(bn256.G1).Add(Key[i], new(bn256.G1).ScalarMult(PKs[i], VSS_SK.Shares[i]))
	}
	//fmt.Printf("CKey is %v\n", CKeys)
	var ckeys []contract.VerificationG1Point
	for i := 0; i < numShares; i++ {
		g1Point = G1ToG1Point(CKeys[i])
		ckeys = append(ckeys, g1Point)
	}
	auth6 := utils.Transact(client, privatekey, big.NewInt(0))
	tx6, _ := Contract.UploadCKeys(auth6, ckeys)
	receipt6, err := bind.WaitMined(context.Background(), client, tx6)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("Upload CKeys Gas used: %d\n", receipt6.GasUsed)
	//Data owner generates a set of DLEQProof prfs_s and publishes the prfs_s on the blockchain

	mul_G := make([]*bn256.G1, numShares)
	mul_H := make([]*bn256.G1, numShares)
	mul_XG := make([]*bn256.G1, numShares)
	mul_XH := make([]*bn256.G1, numShares)
	mul_X := make([]*big.Int, numShares)

	for i := 0; i < numShares; i++ {
		mul_G[i] = g
		mul_H[i] = new(bn256.G1).Add(C.C0, PKs[i])
		mul_X[i] = VSS_SK.Shares[i]
		mul_XG[i] = new(bn256.G1).ScalarMult(g, mul_X[i])
		mul_XH[i] = new(bn256.G1).ScalarMult(mul_H[i], mul_X[i])
	}
	mul_C, mul_Z, mul_XG, mul_XH, mul_RG, mul_RH, _ := dleq.Mul_NewDLEQProof(mul_G, mul_H, mul_XG, mul_XH, mul_X)

	c := make([]*big.Int, numShares)
	z := make([]*big.Int, numShares)
	xG := make([]*bn256.G1, numShares)
	xH := make([]*bn256.G1, numShares)
	rG := make([]*bn256.G1, numShares)
	rH := make([]*bn256.G1, numShares)
	for i := 0; i < numShares; i++ {
		c[i] = mul_C[i]
		z[i] = mul_Z[i]
		xG[i] = mul_XG[i]
		xH[i] = mul_XH[i]
		rG[i] = mul_RG[i]
		rH[i] = mul_RH[i]
	}
	prfs_s := DLEQProofs{C: c, Z: z, XG: xG, XH: xH, RG: rG, RH: rH}

	var Proof_g []contract.VerificationG1Point
	var Proof_gx []contract.VerificationG1Point
	var Proof_h []contract.VerificationG1Point
	var Proof_hx []contract.VerificationG1Point
	Proof_c := make([]*big.Int, numShares)
	var Proof_gr []contract.VerificationG1Point
	var Proof_hr []contract.VerificationG1Point
	Proof_z := make([]*big.Int, numShares)
	for i := 0; i < numShares; i++ {
		g1Point = G1ToG1Point(mul_G[i])
		Proof_g = append(Proof_g, g1Point)
		g1Point = G1ToG1Point(prfs_s.XG[i])
		Proof_gx = append(Proof_gx, g1Point)
		g1Point = G1ToG1Point(mul_H[i])
		Proof_h = append(Proof_h, g1Point)
		g1Point = G1ToG1Point(prfs_s.XH[i])
		Proof_hx = append(Proof_hx, g1Point)
		g1Point = G1ToG1Point(prfs_s.RG[i])
		Proof_gr = append(Proof_gr, g1Point)
		g1Point = G1ToG1Point(prfs_s.RH[i])
		Proof_hr = append(Proof_hr, g1Point)
		Proof_c[i] = prfs_s.C[i]
		Proof_z[i] = prfs_s.Z[i]
	}
	auth7 := utils.Transact(client, privatekey, big.NewInt(0))
	tx7, _ := Contract.UploadDLEQProofCKeys(auth7, Proof_c, Proof_gr, Proof_hr, Proof_z)
	receipt7, err := bind.WaitMined(context.Background(), client, tx7)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("Upload the DLEQ proofs(prfs_s) Gas used: %d\n", receipt7.GasUsed)

	//Data owner generates a set of DLEQProof prfs'_s and publishes the prfs'_s on the blockchain
	_mul_H := make([]*bn256.G1, numShares)
	_mul_XH := make([]*bn256.G1, numShares)
	for i := 0; i < numShares; i++ {
		_mul_H[i] = C.C0
		_mul_XH[i] = new(bn256.G1).ScalarMult(_mul_H[i], mul_X[i])
	}
	mul_C, mul_Z, mul_XG, mul_XH, mul_RG, mul_RH, _ = dleq.Mul_NewDLEQProof(mul_G, _mul_H, mul_XG, _mul_XH, mul_X)

	c = make([]*big.Int, numShares)
	z = make([]*big.Int, numShares)
	xG = make([]*bn256.G1, numShares)
	xH = make([]*bn256.G1, numShares)
	rG = make([]*bn256.G1, numShares)
	rH = make([]*bn256.G1, numShares)
	for i := 0; i < numShares; i++ {
		c[i] = mul_C[i]
		z[i] = mul_Z[i]
		xG[i] = mul_XG[i]
		xH[i] = mul_XH[i]
		rG[i] = mul_RG[i]
		rH[i] = mul_RH[i]
	}
	_prfs_s := DLEQProofs{C: c, Z: z, XG: xG, XH: xH, RG: rG, RH: rH}

	var _Proof_g []contract.VerificationG1Point
	var _Proof_gx []contract.VerificationG1Point
	var _Proof_h []contract.VerificationG1Point
	var _Proof_hx []contract.VerificationG1Point
	_Proof_c := make([]*big.Int, numShares)
	var _Proof_gr []contract.VerificationG1Point
	var _Proof_hr []contract.VerificationG1Point
	_Proof_z := make([]*big.Int, numShares)
	for i := 0; i < numShares; i++ {
		g1Point = G1ToG1Point(mul_G[i])
		_Proof_g = append(_Proof_g, g1Point)
		g1Point = G1ToG1Point(prfs_s.XG[i])
		_Proof_gx = append(_Proof_gx, g1Point)
		g1Point = G1ToG1Point(mul_H[i])
		_Proof_h = append(_Proof_h, g1Point)
		g1Point = G1ToG1Point(prfs_s.XH[i])
		_Proof_hx = append(_Proof_hx, g1Point)
		g1Point = G1ToG1Point(prfs_s.RG[i])
		_Proof_gr = append(_Proof_gr, g1Point)
		g1Point = G1ToG1Point(prfs_s.RH[i])
		_Proof_hr = append(_Proof_hr, g1Point)
		_Proof_c[i] = prfs_s.C[i]
		_Proof_z[i] = prfs_s.Z[i]
	}
	auth8 := utils.Transact(client, privatekey, big.NewInt(0))
	tx8, _ := Contract.UploadDLEQProofKeys(auth8, _Proof_c, _Proof_gr, _Proof_hr, _Proof_z)
	receipt8, err := bind.WaitMined(context.Background(), client, tx8)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("Upload the DLEQ proofs(prfs'_s) Gas used: %d\n", receipt8.GasUsed)

	// //---------------------------------------Figure7 Test-------------------------------------------//
	// //TODO(Figure 7):Test the time cost of Secret-Hiding with the change of TTPs' number(1 THEGEncrypt+1 THEGKeyGen+ 2n DLEQProof)
	// //Randomly generate a plaintext m
	// m, _ := rand.Int(rand.Reader, order)
	// var C *ThresholdElGamal.C
	// var VSS_SK *vss.SecretSharing
	// var Key []*bn256.G1
	// CKeys := make([]*bn256.G1, numShares)

	// g1 := new(bn256.G1)
	// g1Scalar := big.NewInt(1)
	// g := g1.ScalarBaseMult(g1Scalar)

	// mul_G := make([]*bn256.G1, numShares)
	// mul_H := make([]*bn256.G1, numShares)
	// mul_XG := make([]*bn256.G1, numShares)
	// mul_XH := make([]*bn256.G1, numShares)
	// mul_X := make([]*big.Int, numShares)
	// mul_C := make([]*big.Int, numShares)
	// mul_Z := make([]*big.Int, numShares)
	// mul_RG := make([]*bn256.G1, numShares)
	// mul_RH := make([]*bn256.G1, numShares)

	// _mul_H := make([]*bn256.G1, numShares)
	// _mul_XH := make([]*bn256.G1, numShares)
	// _mul_C := make([]*big.Int, numShares)
	// _mul_Z := make([]*big.Int, numShares)
	// _mul_RG := make([]*bn256.G1, numShares)
	// _mul_RH := make([]*bn256.G1, numShares)

	// starttime := time.Now().UnixMilli()
	// for i := 0; i < int(n); i++ {
	// 	//Data owner encrypts the plaintext to ciphertext C and C is published on the blockchain
	// 	C = ThresholdElGamal.THEGEncrypt(m, pko)

	// 	VSS_SK, Key = ThresholdElGamal.THEGKenGen(C, sko, numShares, threshold)

	// 	for i := 0; i < numShares; i++ {
	// 		CKeys[i] = new(bn256.G1).Add(Key[i], new(bn256.G1).ScalarMult(PKs[i], VSS_SK.Shares[i]))
	// 	}

	// 	for i := 0; i < numShares; i++ {
	// 		mul_G[i] = g
	// 		mul_H[i] = new(bn256.G1).Add(C.C0, PKs[i])
	// 		mul_X[i] = VSS_SK.Shares[i]
	// 		mul_XG[i] = new(bn256.G1).ScalarMult(g, mul_X[i])
	// 		mul_XH[i] = new(bn256.G1).ScalarMult(mul_H[i], mul_X[i])
	// 	}

	// 	mul_C, mul_Z, mul_XG, mul_XH, mul_RG, mul_RH, _ = dleq.Mul_NewDLEQProof(mul_G, mul_H, mul_XG, mul_XH, mul_X)

	// 	for i := 0; i < numShares; i++ {
	// 		_mul_H[i] = C.C0
	// 		_mul_XH[i] = new(bn256.G1).ScalarMult(_mul_H[i], mul_X[i])
	// 	}

	// 	_mul_C, _mul_Z, mul_XG, _mul_XH, _mul_RG, _mul_RH, _ = dleq.Mul_NewDLEQProof(mul_G, _mul_H, mul_XG, _mul_XH, mul_X)
	// }
	// endtime := time.Now().UnixMilli()
	// fmt.Printf("Figure 7: Secret-Hiding time cost %d ms\n", (endtime-starttime)/n)

	// fmt.Printf("The ciphertext C is %s\n", C)
	// auth3 := utils.Transact(client, privatekey, big.NewInt(0))
	// tx3, _ := Contract.UploadCiphertext(auth3, G1ToG1Point(C.C0), G1ToG1Point(C.C1))
	// receipt3, err := bind.WaitMined(context.Background(), client, tx3)
	// if err != nil {
	// 	log.Fatalf("Tx receipt failed: %v", err)
	// }
	// fmt.Printf("upload Ciphertext C Gas used: %d\n", receipt3.GasUsed)
	// //Generate the PVSS shares Key of sko, the share commitment Commitments and {g^si} and publish Commitments and {g^si} on the blockchain
	// //TODO(Figure 4):test the time cost of THEGKenGen algorithm with the change of TTPs' number

	// var Gs []contract.VerificationG1Point
	// var Commitments []contract.VerificationG1Point
	// for i := 0; i < numShares; i++ {
	// 	g1Point = G1ToG1Point(VSS_SK.Gs[i])
	// 	Gs = append(Gs, g1Point)
	// }
	// for i := 0; i < threshold; i++ {
	// 	g1Point = G1ToG1Point(VSS_SK.Commitments[i])
	// 	Commitments = append(Commitments, g1Point)
	// }
	// auth4 := utils.Transact(client, privatekey, big.NewInt(0))
	// tx4, _ := Contract.GsAndCommitment(auth4, Gs, Commitments)
	// receipt4, err := bind.WaitMined(context.Background(), client, tx4)
	// if err != nil {
	// 	log.Fatalf("Tx receipt failed: %v", err)
	// }
	// fmt.Printf("upload Gs and Commitments Gas used: %d\n", receipt4.GasUsed)
	// //Data owner uses the TTPs' public keys to encrypt Key to CKey and publishes CKey on the blockchain

	// //fmt.Printf("CKey is %v\n", CKeys)
	// var ckeys []contract.VerificationG1Point
	// for i := 0; i < numShares; i++ {
	// 	g1Point = G1ToG1Point(CKeys[i])
	// 	ckeys = append(ckeys, g1Point)
	// }
	// auth8 := utils.Transact(client, privatekey, big.NewInt(0))
	// tx8, _ := Contract.UploadCKey(auth8, ckeys)
	// receipt8, err := bind.WaitMined(context.Background(), client, tx8)
	// if err != nil {
	// 	log.Fatalf("Tx receipt failed: %v", err)
	// }
	// fmt.Printf("upload CKeys Gas used: %d\n", receipt8.GasUsed)
	// //Data owner generates a set of DLEQProof prfs_s and publishes the prfs_s on the blockchain

	// c := make([]*big.Int, numShares)
	// z := make([]*big.Int, numShares)
	// xG := make([]*bn256.G1, numShares)
	// xH := make([]*bn256.G1, numShares)
	// rG := make([]*bn256.G1, numShares)
	// rH := make([]*bn256.G1, numShares)
	// for i := 0; i < numShares; i++ {
	// 	c[i] = mul_C[i]
	// 	z[i] = mul_Z[i]
	// 	xG[i] = mul_XG[i]
	// 	xH[i] = mul_XH[i]
	// 	rG[i] = mul_RG[i]
	// 	rH[i] = mul_RH[i]
	// }
	// prfs_s := DLEQProofs{C: c, Z: z, XG: xG, XH: xH, RG: rG, RH: rH}

	// var Proof_g []contract.VerificationG1Point
	// var Proof_gx []contract.VerificationG1Point
	// var Proof_h []contract.VerificationG1Point
	// var Proof_hx []contract.VerificationG1Point
	// Proof_c := make([]*big.Int, numShares)
	// var Proof_gr []contract.VerificationG1Point
	// var Proof_hr []contract.VerificationG1Point
	// Proof_z := make([]*big.Int, numShares)
	// for i := 0; i < numShares; i++ {
	// 	g1Point = G1ToG1Point(mul_G[i])
	// 	Proof_g = append(Proof_g, g1Point)
	// 	g1Point = G1ToG1Point(prfs_s.XG[i])
	// 	Proof_gx = append(Proof_gx, g1Point)
	// 	g1Point = G1ToG1Point(mul_H[i])
	// 	Proof_h = append(Proof_h, g1Point)
	// 	g1Point = G1ToG1Point(prfs_s.XH[i])
	// 	Proof_hx = append(Proof_hx, g1Point)
	// 	g1Point = G1ToG1Point(prfs_s.RG[i])
	// 	Proof_gr = append(Proof_gr, g1Point)
	// 	g1Point = G1ToG1Point(prfs_s.RH[i])
	// 	Proof_hr = append(Proof_hr, g1Point)
	// 	Proof_c[i] = prfs_s.C[i]
	// 	Proof_z[i] = prfs_s.Z[i]
	// }
	// auth9 := utils.Transact(client, privatekey, big.NewInt(0))
	// tx9, _ := Contract.UploadDLEQProof(auth9, Proof_g, Proof_gx, Proof_h, Proof_hx, Proof_c, Proof_gr, Proof_hr, Proof_z)
	// receipt9, err := bind.WaitMined(context.Background(), client, tx9)
	// if err != nil {
	// 	log.Fatalf("Tx receipt failed: %v", err)
	// }
	// fmt.Printf("Upload the DLEQ proofs(prfs_s) Gas used: %d\n", receipt9.GasUsed)

	// //Data owner generates a set of DLEQProof prfs'_s and publishes the prfs'_s on the blockchain

	// c = make([]*big.Int, numShares)
	// z = make([]*big.Int, numShares)
	// xG = make([]*bn256.G1, numShares)
	// xH = make([]*bn256.G1, numShares)
	// rG = make([]*bn256.G1, numShares)
	// rH = make([]*bn256.G1, numShares)
	// for i := 0; i < numShares; i++ {
	// 	c[i] = _mul_C[i]
	// 	z[i] = _mul_Z[i]
	// 	xG[i] = mul_XG[i]
	// 	xH[i] = _mul_XH[i]
	// 	rG[i] = _mul_RG[i]
	// 	rH[i] = _mul_RH[i]
	// }
	// _prfs_s := DLEQProofs{C: c, Z: z, XG: xG, XH: xH, RG: rG, RH: rH}

	// var _Proof_g []contract.VerificationG1Point
	// var _Proof_gx []contract.VerificationG1Point
	// var _Proof_h []contract.VerificationG1Point
	// var _Proof_hx []contract.VerificationG1Point
	// _Proof_c := make([]*big.Int, numShares)
	// var _Proof_gr []contract.VerificationG1Point
	// var _Proof_hr []contract.VerificationG1Point
	// _Proof_z := make([]*big.Int, numShares)
	// for i := 0; i < numShares; i++ {
	// 	g1Point = G1ToG1Point(mul_G[i])
	// 	_Proof_g = append(_Proof_g, g1Point)
	// 	g1Point = G1ToG1Point(prfs_s.XG[i])
	// 	_Proof_gx = append(_Proof_gx, g1Point)
	// 	g1Point = G1ToG1Point(mul_H[i])
	// 	_Proof_h = append(_Proof_h, g1Point)
	// 	g1Point = G1ToG1Point(prfs_s.XH[i])
	// 	_Proof_hx = append(_Proof_hx, g1Point)
	// 	g1Point = G1ToG1Point(prfs_s.RG[i])
	// 	_Proof_gr = append(_Proof_gr, g1Point)
	// 	g1Point = G1ToG1Point(prfs_s.RH[i])
	// 	_Proof_hr = append(_Proof_hr, g1Point)
	// 	_Proof_c[i] = prfs_s.C[i]
	// 	_Proof_z[i] = prfs_s.Z[i]
	// }
	// auth10 := utils.Transact(client, privatekey, big.NewInt(0))
	// tx10, _ := Contract.UploadDLEQProof(auth10, _Proof_g, _Proof_gx, _Proof_h, _Proof_hx, _Proof_c, _Proof_gr, _Proof_hr, _Proof_z)
	// receipt10, err := bind.WaitMined(context.Background(), client, tx10)
	// if err != nil {
	// 	log.Fatalf("Tx receipt failed: %v", err)
	// }
	// fmt.Printf("Upload the DLEQ proofs(prfs'_s) Gas used: %d\n", receipt10.GasUsed)

	//--------------------------Key-Verification-----------------------------//
	//The verification of {g^si}(finish on the blockchain)
	result := vss.VerifyShare(VSS_SK.Gs, VSS_SK.Commitments)
	fmt.Printf("The off-chain result of VSSVerify is %v\n", result)
	// arr := make([]*big.Int, numShares*3+threshold*3)

	// for i := 0; i < numShares; i++ {
	// 	arr[i] = big.NewInt(int64(i + 1))
	// }

	// for i := numShares; i < 3*numShares; i = i + 2 {
	// 	_gs := G1ToBigIntArray(VSS_SK.Gs[(i-numShares)/2])
	// 	arr[i] = _gs[0]
	// 	arr[i+1] = _gs[1]
	// }

	// for i := 3 * numShares; i < 3*numShares+threshold; i++ {
	// 	arr[i] = big.NewInt(int64(i - 3*numShares))
	// }

	// for i := 3*numShares + threshold; i < 3*numShares+3*threshold; i = i + 2 {
	// 	_gs := G1ToBigIntArray(VSS_SK.Commitments[(i-(3*numShares+threshold))/2])
	// 	arr[i] = _gs[0]
	// 	arr[i+1] = _gs[1]
	// }

	// fmt.Printf("The converted set is %v\n", arr)

	auth9 := utils.Transact(client, privatekey, big.NewInt(0))
	tx9, _ := Contract.VSSVerify(auth9, Commitments)
	VSSResult, _ := Contract.GetVrfResult(&bind.CallOpts{})
	receipt9, err := bind.WaitMined(context.Background(), client, tx9)

	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}

	fmt.Printf("VSSVerify Result: %v\n", VSSResult)
	fmt.Printf("VSSVerify Gas used: %d\n", receipt9.GasUsed)

	Error := dleq.Mul_Verify(prfs_s.C, prfs_s.Z, mul_G, mul_H, prfs_s.XG, prfs_s.XH, prfs_s.RG, prfs_s.RH)
	fmt.Printf("The off-chain result of DLEQVrf(prfs_s) is %v\n", Error)
	auth10 := utils.Transact(client, privatekey, big.NewInt(0))
	tx10, _ := Contract.DLEQVerifyCKeys(auth10)
	DLEQResult, _ := Contract.GetVrfResult(&bind.CallOpts{})
	receipt10, err := bind.WaitMined(context.Background(), client, tx10)

	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}

	fmt.Printf("DLEQVrf Gas used: %d\n", receipt10.GasUsed)
	fmt.Printf("DLEQVrf result is %v\n", DLEQResult)
	//TODO(Figure 8): Test the gas comsuption of Key-Verification with the change of TTPs' numbers(n VSSVerify+n DLEQVrf)
	fmt.Printf("Figure 8: the key-verification Gas Used %v\n", receipt9.GasUsed+receipt10.GasUsed)
	//-------------------------------Key-Delegation-------------------------------------//
	//TTPs' use their private keys SKs to decrypt CKey to TTPs_Key
	//TODO(Figure 9):Test the time cost of Key-Delagation with the change of TTPs' number(n decryption operations and n EGEncrypt)
	TTPs_Key := make([]*bn256.G1, numShares)
	var EKeys *ElGamal.EK
	starttime = time.Now().UnixMilli()

	for i := 0; i < int(n); i++ {
		for i := 0; i < numShares; i++ {
			TTPs_Key[i] = new(bn256.G1).Add(CKeys[i], new(bn256.G1).Neg(new(bn256.G1).ScalarMult(VSS_SK.Gs[i], SKs[i])))
		}
		//TTPs use the public key pku to encrypts TTPs_Key to EKey and the EKey is published on the blockchain
		EKeys = ElGamal.EGEncrypt(TTPs_Key, pku, numShares)
	}

	endtime = time.Now().UnixMilli()
	fmt.Printf("Figure 9: the time cost of key-delegation is %v ms\n", (endtime-starttime)/n)
	//fmt.Printf("The key encrypted by TTPs is %v\n", EKeys)
	var ekeys0 []contract.VerificationG1Point
	var ekeys1 []contract.VerificationG1Point

	for i := 0; i < numShares; i++ {
		g1Point = G1ToG1Point(EKeys.EK0[i])
		ekeys0 = append(ekeys0, g1Point)
		g1Point = G1ToG1Point(EKeys.EK1[i])
		ekeys1 = append(ekeys1, g1Point)
	}

	auth11 := utils.Transact(client, privatekey, big.NewInt(0))
	tx11, _ := Contract.UploadEKeys(auth11, ekeys0, ekeys1)
	receipt11, err := bind.WaitMined(context.Background(), client, tx11)

	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}

	fmt.Printf("Upload EKeys Gas used: %d\n", receipt11.GasUsed)
	//---------------------------------------Secret-Recovery-----------------------------------------//
	//Data user uses their private keys sku to decrypt TTPs_Key to _Key
	_Key := make([]*bn256.G1, numShares)
	var _m *bn256.G1
	starttime = time.Now().UnixMilli()

	for i := 0; i < int(n); i++ {
		_Key = ElGamal.EGDecrypt(EKeys, sku, numShares)

		for i := 0; i < numShares; i++ {
			_prfs_s.XH[i] = _Key[i]
		}
		//Data user verifies the _Key
		Error = dleq.Mul_Verify(_prfs_s.C, _prfs_s.Z, mul_G, _mul_H, _prfs_s.XG, _prfs_s.XH, _prfs_s.RG, _prfs_s.RH)
		//fmt.Printf("The result of DLEQVrf(prfs'_s) is %v\n", Error)
		//Data user decrypts the ciphertext to plaintext _m
		KeyIndices := make([]*big.Int, threshold)
		for i := 0; i < threshold; i++ {
			KeyIndices[i] = big.NewInt(int64(i + 1))
		}
		//TODO(Figure 5):Test the time cost of THEGDecrypt algorithm with the change of TTPs' number
		_m = ThresholdElGamal.THEGDecrypt(C, _Key, KeyIndices, threshold)
	}

	endtime = time.Now().UnixMilli()
	fmt.Printf("Figure 10: the time cost of Secret-Recovery is %v ms\n", (endtime-starttime)/n)
	fmt.Printf("The plaintext _m is %s\n", _m)
	//---------------------------------------Dispute----------------------------------------//
	//Data user generates a DIS and publishes it on the blockchain(e.g. _Key[0])
	numDispute := 1 //the number of dispute
	DIS := make([]*bn256.G1, numDispute)
	DIS[0] = new(bn256.G1).ScalarMult(EKeys.EK0[0], sku)
	auth12 := utils.Transact(client, privatekey, big.NewInt(0))
	tx12, _ := Contract.UploadDispute(auth12, G1ToG1Point(DIS[0]))
	receipt12, err := bind.WaitMined(context.Background(), client, tx12)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}

	fmt.Printf("Upload a dispute Gas used: %d\n", receipt12.GasUsed)
	_xG := new(bn256.G1).ScalarMult(g, sku)
	_xH := new(bn256.G1).ScalarMult(EKeys.EK0[0], sku)
	//Data user generates the DLEQProof of sku prfs_sku and publishes the prfs_sku on the blockchain
	starttime = time.Now().UnixMicro()
	var prfs_sku DLEQProof

	for i := 0; i < int(n); i++ {
		_c, _z, _rG, _rH, _ := dleq.NewDLEQProof(g, EKeys.EK0[0], _xG, _xH, sku)
		prfs_sku = DLEQProof{C: _c, Z: _z, XG: _xG, XH: _xH, RG: _rG, RH: _rH}
	}

	endtime = time.Now().UnixMicro()
	fmt.Printf("DLEQ time cost %d us\n", (endtime-starttime)/n)
	Dis_proof_c := prfs_sku.C
	Dis_proof_gr := G1ToG1Point(prfs_sku.RG)
	Dis_proof_hr := G1ToG1Point(prfs_sku.RH)
	Dis_proof_z := prfs_sku.Z

	auth13 := utils.Transact(client, privatekey, big.NewInt(0))
	tx13, _ := Contract.UploadDisputeProof(auth13, Dis_proof_c, Dis_proof_gr, Dis_proof_hr, Dis_proof_z)
	receipt13, err := bind.WaitMined(context.Background(), client, tx13)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}

	fmt.Printf("Upload a disputeDLEQProof(prfs_sku) Gas used: %d\n", receipt13.GasUsed)
	//Vefify the dispute DIS
	starttime = time.Now().UnixMicro()

	for i := 0; i < int(n); i++ {
		Error = dleq.Verify(prfs_sku.C, prfs_sku.Z, g, EKeys.EK0[0], pku, DIS[0], prfs_sku.RG, prfs_sku.RH)
	}

	endtime = time.Now().UnixMicro()
	fmt.Printf("DLEQVrf time cost %d us\n", (endtime-starttime)/n)
	fmt.Printf("The off-chain result of dispute verification is %v\n", Error)
	fmt.Printf("Disproof g is %v\n", g)
	auth14 := utils.Transact(client, privatekey, big.NewInt(0))
	tx14, _ := Contract.DLEQVerifyDis(auth14, big.NewInt(0))
	receipt14, err := bind.WaitMined(context.Background(), client, tx14)

	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}

	//TODO(Figure 11):Test the gas cost of a dispute verification
	fmt.Printf("Figure 11: Dispute verification Gas used %v\n", receipt14.GasUsed)

}
