package main

import (
	"context"
	"crypto/rand"
	"dttp/compile/contract"
	"dttp/crypto/ElGamal"
	"dttp/crypto/Threshold_ElGamal"
	"dttp/crypto/dleq"
	"dttp/crypto/vss"
	"dttp/utils"
	"fmt"
	"log"
	"math/big"

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

	auth0 := utils.Transact(client, privatekey, big.NewInt(0))

	address, tx0 := utils.Deploy(client, contract_name, auth0)

	receipt, err := bind.WaitMined(context.Background(), client, tx0)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("Deploy Gas used: %d\n", receipt.GasUsed)

	Contract, err := contract.NewContract(common.HexToAddress(address.Hex()), client)
	if err != nil {
		fmt.Println(err)
	}

	// the number of key shares
	numShares := 5
	// threshold value
	threshold := 3

	//------------------------------------------Registration-------------------------------------//
	//TODO(Figure 6)： test the gas comsuption of uploading TTPs' PKs with the number challenge of TTPs
	//Data owner's key pair (sko,pko) and the public key pko is published on the blockchain
	sko, pko := Threshold_ElGamal.THEGSetup()
	fmt.Printf("sku为：%v\n", sko)
	fmt.Printf("pku为：%v\n", pko)
	//TTPs' key pairs (SKs, PKs) and these public keys PKs are published on the blockchain
	SKs := make([]*big.Int, numShares)  //the set of TTPs' private key
	PKs := make([]*bn256.G1, numShares) //the set of TTPs' public key
	TTPs_PKs := make([][2]*big.Int, numShares)
	for i := 0; i < numShares; i++ {
		sk, pk, _ := bn256.RandomG1(rand.Reader)
		SKs[i] = sk
		PKs[i] = pk
		TTPs_PKs[i] = G1ToBigIntArray(pk)
	}
	auth2 := utils.Transact(client, privatekey, big.NewInt(0))
	tx2, _ := Contract.UploadMultipleTTPPk(auth2, TTPs_PKs)
	receipt2, err := bind.WaitMined(context.Background(), client, tx2)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("upload TTP 's pk Gas used: %d\n", receipt2.GasUsed)
	//Data user's key pair and the public key is published on the blockchain
	sku, pku := ElGamal.EGSetup()
	fmt.Printf("sku为：%v\n", sku)
	fmt.Printf("pku为：%v\n", pku)
	//---------------------------------------Secret-Hiding-----------------------------------------//
	//Randomly generate a plaintext m
	m, _ := rand.Int(rand.Reader, order)
	//Data owner encrypts the plaintext to ciphertext C and C is published on the blockchain
	C := Threshold_ElGamal.THEGEncrypt(m, pko)
	fmt.Printf("The ciphertext C is %s\n", C)

	auth3 := utils.Transact(client, privatekey, big.NewInt(0))
	tx3, _ := Contract.UploadCiphertext(auth3, G1ToBigIntArray(C.C0), G1ToBigIntArray(C.C1))
	receipt3, err := bind.WaitMined(context.Background(), client, tx3)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("upload Ciphertext C Gas used: %d\n", receipt3.GasUsed)
	//Generate the PVSS shares Key of sko, the share commitment Commitments and {g^si} and publish Commitments and {g^si} on the blockchain
	VSS_SK, Key := Threshold_ElGamal.THEGKenGen(C, sko, numShares, threshold)
	//fmt.Printf("The Gs is %v\n", VSS_SK.Gs)
	//fmt.Printf("The commitments is %v\n", VSS_SK.Commitments)
	Gs := make([][2]*big.Int, numShares)
	Commitments := make([][2]*big.Int, threshold)
	for i := 0; i < numShares; i++ {
		Gs[i] = G1ToBigIntArray(VSS_SK.Gs[i])
	}
	for i := 0; i < threshold; i++ {
		Commitments[i] = G1ToBigIntArray(VSS_SK.Commitments[i])
	}
	auth4 := utils.Transact(client, privatekey, big.NewInt(0))
	tx4, _ := Contract.GsAndCommitment(auth4, Gs, Commitments)
	receipt4, err := bind.WaitMined(context.Background(), client, tx4)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("upload Gs and Commitments Gas used: %d\n", receipt4.GasUsed)
	//Data owner uses the TTPs' public keys to encrypt Key to CKey and publishes CKey on the blockchain
	CKeys := make([]*bn256.G1, numShares)
	for i := 0; i < numShares; i++ {
		CKeys[i] = new(bn256.G1).Add(Key[i], new(bn256.G1).ScalarMult(PKs[i], VSS_SK.Shares[i]))
	}
	fmt.Printf("CKey is %v\n", CKeys)
	ckeys := make([][2]*big.Int, numShares)
	for i := 0; i < numShares; i++ {
		ckeys[i] = G1ToBigIntArray(CKeys[i])
	}
	auth8 := utils.Transact(client, privatekey, big.NewInt(0))
	tx8, _ := Contract.UploadCKey(auth8, ckeys)
	receipt8, err := bind.WaitMined(context.Background(), client, tx8)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("upload CKeys Gas used: %d\n", receipt8.GasUsed)
	//Data owner generates a set of DLEQProof prfs_s and publishes the prfs_s on the blockchain
	g1 := new(bn256.G1)
	g1Scalar := big.NewInt(1)
	g := g1.ScalarBaseMult(g1Scalar)

	mul_G := make([]*bn256.G1, numShares)
	mul_H := make([]*bn256.G1, numShares)
	mul_X := make([]*big.Int, numShares)

	for i := 0; i < numShares; i++ {
		mul_G[i] = g
		mul_H[i] = new(bn256.G1).Add(C.C0, PKs[i])
		mul_X[i] = VSS_SK.Shares[i]
	}
	mul_C, mul_Z, mul_XG, mul_XH, mul_RG, mul_RH, _ := dleq.Mul_NewDLEQProof(mul_G, mul_H, mul_X)

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

	Proof_g := make([][2]*big.Int, numShares)
	Proof_gx := make([][2]*big.Int, numShares)
	Proof_h := make([][2]*big.Int, numShares)
	Proof_hx := make([][2]*big.Int, numShares)
	Proof_c := make([]*big.Int, numShares)
	Proof_gr := make([][2]*big.Int, numShares)
	Proof_hr := make([][2]*big.Int, numShares)
	Proof_z := make([]*big.Int, numShares)
	for i := 0; i < numShares; i++ {
		Proof_g[i] = G1ToBigIntArray(mul_G[i])
		Proof_gx[i] = G1ToBigIntArray(prfs_s.XG[i])
		Proof_h[i] = G1ToBigIntArray(mul_H[i])
		Proof_hx[i] = G1ToBigIntArray(prfs_s.XH[i])
		Proof_gr[i] = G1ToBigIntArray(prfs_s.RG[i])
		Proof_hr[i] = G1ToBigIntArray(prfs_s.RH[i])
		Proof_c[i] = prfs_s.C[i]
		Proof_z[i] = prfs_s.Z[i]
	}
	auth9 := utils.Transact(client, privatekey, big.NewInt(0))
	tx9, _ := Contract.UploadDLEQProof(auth9, Proof_g, Proof_gx, Proof_h, Proof_hx, Proof_c, Proof_gr, Proof_hr, Proof_z)
	receipt9, err := bind.WaitMined(context.Background(), client, tx9)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}

	fmt.Printf("Upload the DLEQ proofs(prfs_s) Gas used: %d\n", receipt9.GasUsed)

	//Data owner generates a set of DLEQProof prfs'_s and publishes the prfs'_s on the blockchain
	_mul_H := make([]*bn256.G1, numShares)
	for i := 0; i < numShares; i++ {
		_mul_H[i] = C.C0
	}
	mul_C, mul_Z, mul_XG, mul_XH, mul_RG, mul_RH, _ = dleq.Mul_NewDLEQProof(mul_G, _mul_H, mul_X)

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
	fmt.Printf("The DLEQProof(prfs'_s) is %v\n", _prfs_s)

	_Proof_g := make([][2]*big.Int, numShares)
	_Proof_gx := make([][2]*big.Int, numShares)
	_Proof_h := make([][2]*big.Int, numShares)
	_Proof_hx := make([][2]*big.Int, numShares)
	_Proof_c := make([]*big.Int, numShares)
	_Proof_gr := make([][2]*big.Int, numShares)
	_Proof_hr := make([][2]*big.Int, numShares)
	_Proof_z := make([]*big.Int, numShares)
	for i := 0; i < numShares; i++ {
		_Proof_g[i] = G1ToBigIntArray(mul_G[i])
		_Proof_gx[i] = G1ToBigIntArray(_prfs_s.XG[i])
		_Proof_h[i] = G1ToBigIntArray(_mul_H[i])
		_Proof_hx[i] = G1ToBigIntArray(_prfs_s.XH[i])
		_Proof_gr[i] = G1ToBigIntArray(_prfs_s.RG[i])
		_Proof_hr[i] = G1ToBigIntArray(_prfs_s.RH[i])
		_Proof_c[i] = _prfs_s.C[i]
		_Proof_z[i] = _prfs_s.Z[i]
	}
	auth10 := utils.Transact(client, privatekey, big.NewInt(0))
	tx10, _ := Contract.UploadDLEQProof(auth10, _Proof_g, _Proof_gx, _Proof_h, _Proof_hx, _Proof_c, _Proof_gr, _Proof_hr, _Proof_z)
	receipt10, err := bind.WaitMined(context.Background(), client, tx10)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("Upload the DLEQ proofs(prfs'_s) Gas used: %d\n", receipt10.GasUsed)

	//--------------------------Key-Verification-----------------------------//
	//The verification of {g^si}(finish on the blockchain)
	result := vss.VerifyShare(VSS_SK.Gs, VSS_SK.Commitments)
	fmt.Printf("The off-chain result of VSS.verify is %v\n", result)
	arr := make([]*big.Int, numShares*3+threshold*3)
	for i := 0; i < numShares; i++ {
		arr[i] = big.NewInt(int64(i + 1))
	}
	for i := numShares; i < 3*numShares; i = i + 2 {
		_gs := G1ToBigIntArray(VSS_SK.Gs[(i-numShares)/2])
		arr[i] = _gs[0]
		arr[i+1] = _gs[1]
	}
	for i := 3 * numShares; i < 3*numShares+threshold; i++ {
		arr[i] = big.NewInt(int64(i - 3*numShares))
	}
	for i := 3*numShares + threshold; i < 3*numShares+3*threshold; i = i + 2 {
		_gs := G1ToBigIntArray(VSS_SK.Commitments[(i-(3*numShares+threshold))/2])
		arr[i] = _gs[0]
		arr[i+1] = _gs[1]
	}
	//fmt.Printf("The converted set is %v\n", arr)

	auth5 := utils.Transact(client, privatekey, big.NewInt(0))
	tx5, _ := Contract.VSSVerify(auth5, arr, big.NewInt(int64(numShares)), big.NewInt(int64(threshold)))
	VSSResult, _ := Contract.Get(&bind.CallOpts{})
	receipt5, err := bind.WaitMined(context.Background(), client, tx5)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("VSS.Verify Gas used: %d\n", receipt5.GasUsed)
	fmt.Printf("VSS.Verify Result: %v\n", VSSResult)

	Error := dleq.Mul_Verify(prfs_s.C, prfs_s.Z, mul_G, mul_H, prfs_s.XG, prfs_s.XH, prfs_s.RG, prfs_s.RH)
	fmt.Printf("The off-chain result of DLEQVrf(prfs_s) is %v\n", Error)
	auth6 := utils.Transact(client, privatekey, big.NewInt(0))
	tx6, _ := Contract.MulDELQVerify(auth6, Proof_g, Proof_gx, Proof_h, Proof_hx, Proof_c, Proof_gr, Proof_hr, Proof_z)
	DLEQResult, _ := Contract.Get(&bind.CallOpts{})
	receipt6, err := bind.WaitMined(context.Background(), client, tx6)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("DLEQVerify Gas used: %d\n", receipt6.GasUsed)
	fmt.Printf("DLEQ verification result is %v\n", DLEQResult)

	//-------------------------------Key-Delegation-------------------------------------//
	//TTPs' use their private keys SKs to decrypt CKey to TTPs_Key
	TTPs_Key := make([]*bn256.G1, numShares)
	for i := 0; i < numShares; i++ {
		TTPs_Key[i] = new(bn256.G1).Add(CKeys[i], new(bn256.G1).Neg(new(bn256.G1).ScalarMult(VSS_SK.Gs[i], SKs[i])))
	}
	//TTPs use the public key pku to encrypts TTPs_Key to EKey and the EKey is published on the blockchain
	EKeys := ElGamal.EGEncrypt(TTPs_Key, pku, numShares)
	fmt.Printf("The key encrypted by TTPs is %v\n", EKeys)
	ekeys0 := make([][2]*big.Int, numShares)
	ekeys1 := make([][2]*big.Int, numShares)
	for i := 0; i < numShares; i++ {
		ekey0 := G1ToBigIntArray(EKeys.EK0[i])
		ekey1 := G1ToBigIntArray(EKeys.EK1[i])
		ekeys0[i] = ekey0
		ekeys1[i] = ekey1
	}
	auth11 := utils.Transact(client, privatekey, big.NewInt(0))
	tx11, _ := Contract.UploadEKey(auth11, ekeys0, ekeys1)
	receipt11, err := bind.WaitMined(context.Background(), client, tx11)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("Upload EKeys Gas used: %d\n", receipt11.GasUsed)
	//---------------------------------------Secret-Recovery-----------------------------------------//
	//Data user uses their private keys sku to decrypt TTPs_Key to _Key
	_Key := make([]*bn256.G1, numShares)
	for i := 0; i < numShares; i++ {
		_Key = ElGamal.EGDecrypt(EKeys, sku, numShares)
	}

	for i := 0; i < numShares; i++ {
		_prfs_s.XH[i] = _Key[i]
	}
	//Data user verifies the _Key
	Error = dleq.Mul_Verify(_prfs_s.C, _prfs_s.Z, mul_G, _mul_H, _prfs_s.XG, _prfs_s.XH, _prfs_s.RG, _prfs_s.RH)
	fmt.Printf("The result of DLEQVrf(prfs'_s) is %v\n", Error)
	//Data user decrypts the ciphertext to plaintext _m
	KeyIndices := make([]*big.Int, threshold)
	for i := 0; i < threshold; i++ {
		KeyIndices[i] = big.NewInt(int64(i + 1))
	}
	_m := Threshold_ElGamal.THEGDecrypt(C, _Key, KeyIndices, threshold)
	fmt.Printf("The plaintext _m is %s\n", _m)
	//---------------------------------------Dispute----------------------------------------//
	//Data user generates a DIS and publishes it on the blockchain(e.g. _Key[0])
	numDispute := 1 //the number of dispute
	DIS := make([]*bn256.G1, numDispute)
	DIS[0] = new(bn256.G1).ScalarMult(EKeys.EK0[0], sku)
	//Data user generates the DLEQProof of sku prfs_sku and publishes the prfs_sku on the blockchain
	_c, _z, _xG, _xH, _rG, _rH, _ := dleq.NewDLEQProof(g, EKeys.EK0[0], sku)
	prfs_sku := DLEQProof{C: _c, Z: _z, XG: _xG, XH: _xH, RG: _rG, RH: _rH}
	Dis_proof_g := make([][2]*big.Int, 1)
	Dis_proof_gx := make([][2]*big.Int, 1)
	Dis_proof_h := make([][2]*big.Int, 1)
	Dis_proof_hx := make([][2]*big.Int, 1)
	Dis_proof_c := make([]*big.Int, 1)
	Dis_proof_gr := make([][2]*big.Int, 1)
	Dis_proof_hr := make([][2]*big.Int, 1)
	Dis_proof_z := make([]*big.Int, 1)
	for i := 0; i < numDispute; i++ {
		Dis_proof_g[i] = G1ToBigIntArray(g)
		Dis_proof_h[i] = G1ToBigIntArray(EKeys.EK0[i])
		Dis_proof_gx[i] = G1ToBigIntArray(prfs_sku.XG)
		Dis_proof_hx[i] = G1ToBigIntArray(prfs_sku.XH)
		Dis_proof_gr[i] = G1ToBigIntArray(prfs_sku.RG)
		Dis_proof_hr[i] = G1ToBigIntArray(prfs_sku.RH)
		Dis_proof_c[i] = prfs_sku.C
		Dis_proof_z[i] = prfs_sku.Z
	}
	auth12 := utils.Transact(client, privatekey, big.NewInt(0))
	tx12, _ := Contract.UploadDLEQProof(auth12, Dis_proof_g, Dis_proof_gx, Dis_proof_h, Dis_proof_hx, Dis_proof_c, Dis_proof_gr, Dis_proof_hr, Dis_proof_z)
	receipt12, err := bind.WaitMined(context.Background(), client, tx12)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("Upload a disputeDLEQProof(prfs_sku) Gas used: %d\n", receipt12.GasUsed)
	//Vefify the dispute DIS
	Error = dleq.Verify(prfs_sku.C, prfs_sku.Z, g, EKeys.EK0[0], pku, DIS[0], prfs_sku.RG, prfs_sku.RH)
	fmt.Printf("The off-chain result of dispute verification is %v\n", Error)
	auth13 := utils.Transact(client, privatekey, big.NewInt(0))
	tx13, _ := Contract.DELQVerify(auth13, Dis_proof_g[0], Dis_proof_gx[0], Dis_proof_h[0], Dis_proof_hx[0], Dis_proof_c[0], Dis_proof_gr[0], Dis_proof_hr[0], Dis_proof_z[0])
	DisputeResult, _ := Contract.Get(&bind.CallOpts{})
	receipt13, err := bind.WaitMined(context.Background(), client, tx13)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}
	fmt.Printf("Verify a dispute Gas used: %d\n", receipt13.GasUsed)
	fmt.Printf("Dispute verification result is %v\n", DisputeResult)
}
