package main

import (
	"fmt"
	"crypto/rand"
	"math/big"
	"dttp/crypto/Threshold_ElGamal"
	"dttp/crypto/ElGamal"
	"dttp/crypto/dleq"
	"dttp/crypto/vss"
	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/google"
)

var order=bn256.Order

type DLEQProofs struct{
	C 	[]*big.Int
	Z 	[]*big.Int
	XG 	[]*bn256.G1
	XH 	[]*bn256.G1
	RG 	[]*bn256.G1
	RH 	[]*bn256.G1
}

type DLEQProof struct{
	C 	*big.Int
	Z 	*big.Int
	XG 	*bn256.G1
	XH 	*bn256.G1
	RG 	*bn256.G1
	RH 	*bn256.G1
}

type EK struct{
	EK0 	[]*bn256.G1
	EK1 	[]*bn256.G1
}

func main(){
//the number of key shares
numShares:=5
//threshold value
threshold:=3
//-------------------Registration-------------------//
	//Data owner's key pair (sko,pko) and the public key pko is published on the blockchain
	sko,pko:=Threshold_ElGamal.THEGSetup()
	
	//TTPs' key pairs (SKs, PKs) and these public keys PKs are published on the blockchain
	SKs :=make([]*big.Int, numShares)  //the set of TTPs' private key
	PKs :=make([]*bn256.G1,numShares) //the set of TTPs' public key
	for i:=0;i<numShares;i++{
		sk,pk,_:=bn256.RandomG1(rand.Reader)
		SKs[i]=sk
		PKs[i]=pk
	}
	//Data user's key pair and the public key is published on the blockchain
	sku,pku:=ElGamal.EGSetup()
//-------------------Secret-Hiding-------------------//
    //Randomly generate a plaintext m
    m,_ := rand.Int(rand.Reader, order)
    //Data owner encrypts the plaintext to ciphertext C and C is published on the blockchain
    C:=Threshold_ElGamal.THEGEncrypt(m,pko)
    fmt.Printf("The ciphertext C is %s\n",C)
    //Generate the PVSS shares Key of sko, the share commitment Commitments and {g^si} and publish Commitments and {g^si} on the blockchain
    VSS_SK,Key:=Threshold_ElGamal.THEGKenGen(C, sko, numShares, threshold)
    
    //Data owner uses the TTPs' public keys to encrypt Key to CKey and publishes CKey on the blockchain
    CKey:=make([]*bn256.G1,numShares)
    for i:=0;i<numShares;i++{
    	CKey[i]=new(bn256.G1).Add(Key[i], new(bn256.G1).ScalarMult(PKs[i],VSS_SK.Shares[i]))
    }
    //Data owner generates a set of DLEQProof prfs_s and publishes the prfs_s on the blockchain
    g1 := new(bn256.G1)
    g1Scalar := big.NewInt(1)
    g:=g1.ScalarBaseMult(g1Scalar)
	
    mul_G := make([]*bn256.G1,numShares)
    mul_H := make([]*bn256.G1, numShares)
    mul_X := make([]*big.Int, numShares)
	
    for i:=0;i<numShares;i++{
	mul_G[i]=g
	mul_H[i]=new(bn256.G1).Add(C.C0,PKs[i])
	mul_X[i]=VSS_SK.Shares[i]
    }
    mul_C, mul_Z, mul_XG, mul_XH, mul_RG, mul_RH, _:= dleq.Mul_NewDLEQProof(mul_G, mul_H, mul_X)
    c:=make([]*big.Int, numShares)
    z:=make([]*big.Int, numShares)
    xG:=make([]*bn256.G1,numShares)
    xH:=make([]*bn256.G1,numShares)
    rG:=make([]*bn256.G1,numShares)
    rH:=make([]*bn256.G1,numShares)
    for i:=0;i<numShares;i++{
    	c[i]=mul_C[i]
    	z[i]=mul_Z[i]
    	xG[i]=mul_XG[i]
    	xH[i]=mul_XH[i]
    	rG[i]=mul_RG[i]
    	rH[i]=mul_RH[i]
    }
    prfs_s:=DLEQProofs {C:c, Z:z, XG:xG, XH:xH, RG:rG, RH:rH} 
    
    //Data owner generates a set of DLEQProof prfs_s and publishes the prfs'_s on the blockchain
    _mul_H := make([]*bn256.G1, numShares)
    for i:=0;i<numShares;i++{
	_mul_H[i]=C.C0
    }
    mul_C, mul_Z, mul_XG, mul_XH, mul_RG, mul_RH, _= dleq.Mul_NewDLEQProof(mul_G, _mul_H, mul_X)
	
    c=make([]*big.Int, numShares)
    z=make([]*big.Int, numShares)
    xG=make([]*bn256.G1,numShares)
    xH=make([]*bn256.G1,numShares)
    rG=make([]*bn256.G1,numShares)
    rH=make([]*bn256.G1,numShares)
    for i:=0;i<numShares;i++{
    	c[i]=mul_C[i]
    	z[i]=mul_Z[i]
    	xG[i]=mul_XG[i]
    	xH[i]=mul_XH[i]
    	rG[i]=mul_RG[i]
    	rH[i]=mul_RH[i]
    }
    _prfs_s:=DLEQProofs {C:c, Z:z, XG:xG, XH:xH, RG:rG, RH:rH}
//-------------------Key-Verification-------------------// 
    //The verification of {g^si}(finish on the blockchain)
    result:=vss.VerifyShare(VSS_SK.Gs,VSS_SK.Commitments)
    fmt.Printf("The result of VSS.verify is %v\n", result)
    //The verification of prfs_s(finish on the blockchain)
    Error:= dleq.Mul_Verify(prfs_s.C, prfs_s.Z, mul_G, mul_H, prfs_s.XG, prfs_s.XH, prfs_s.RG, prfs_s.RH)
    fmt.Printf("The result of DLEQVrf(prfs_s) is %v\n",Error)
//-------------------Key-Delegation-------------------// 
    //TTPs' use their private keys SKs to decrypt CKey to TTPs_Key
    TTPs_Key:=make([]*bn256.G1,numShares) 
    for i:=0;i<numShares;i++{			
	TTPs_Key[i]=new(bn256.G1).Add(CKey[i],new(bn256.G1).Neg(new(bn256.G1).ScalarMult(VSS_SK.Gs[i],SKs[i])))
    } 
    //TTPs use the public key pku to encrypts TTPs_Key to EKey and the EKey is published on the blockchain
    EKey:=ElGamal.EGEncrypt(TTPs_Key,pku,numShares)
//-------------------Key-Delegation-------------------//
    //Data user uses their private keys sku to decrypt TTPs_Key to _Key
    _Key:= make([]*bn256.G1,numShares)
    for i:=0;i<numShares;i++{
	_Key=ElGamal.EGDecrypt(EKey,sku,numShares)
    }

    for i:=0;i<numShares;i++{
	    _prfs_s.XH[i]=_Key[i]
    }
    //Data user verifies the _Key（finish on the blockchain）
    Error= dleq.Mul_Verify(_prfs_s.C, _prfs_s.Z, mul_G, _mul_H, _prfs_s.XG, _prfs_s.XH, _prfs_s.RG, _prfs_s.RH)
    fmt.Printf("The result of DLEQVrf(prfs'_s) is %v\n",Error)
    //Data user decrypts the ciphertext to plaintext _m
    KeyIndices := make([]*big.Int, threshold)
	for i := 0; i < threshold; i++ {
		KeyIndices[i] = big.NewInt(int64(i + 1))
	}
    _m:=Threshold_ElGamal.THEGDecrypt(C, _Key, KeyIndices, threshold)
    fmt.Printf("The plaintext _m is %s\n",_m)
//-------------------Dispute-------------------//
    //Data user generates a DIS and publishes it on the blockchain(e.g. _Key[0])
    DIS:=new(bn256.G1).ScalarMult(EKey.EK0[0],sku)
    //Data user generates the DLEQProof of sku prfs_sku and publishes the prfs_sku on the blockchain
    _c, _z, _xG, _xH, _rG, _rH, _ := dleq.NewDLEQProof(g, EKey.EK0[0], sku)
    prfs_sku:=DLEQProof{C:_c, Z:_z, XG:_xG, XH:_xH, RG:_rG, RH:_rH} 
    //Vefify the dispute DIS 
     _Error:=dleq.Verify(prfs_sku.C, prfs_sku.Z, g, EKey.EK0[0], pku, DIS, prfs_sku.RG, prfs_sku.RH)
    fmt.Printf("The result of dispute verification is %v\n",_Error)
}
