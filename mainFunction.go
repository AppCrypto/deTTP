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
//TTPi的数量
numShares:=5
//门限值
threshold:=3
//-------------------Registration-------------------//
	//Data owner注册密钥对，其中公钥pko公开在链上
	sko,pko:=Threshold_ElGamal.THEGSetup()
	fmt.Printf("Data owner的私钥为：%s\n",sko)
	fmt.Printf("Data owner的公钥为：%s\n",pko)
	
	//TTPs注册密钥对，其中公钥pki公开在链上
	SKs :=make([]*big.Int, numShares)  //TTPs的私钥
	PKs :=make([]*bn256.G1,numShares) //TTPs的公钥
	for i:=0;i<numShares;i++{
		sk,pk,_:=bn256.RandomG1(rand.Reader)
		SKs[i]=sk
		PKs[i]=pk
	}
	fmt.Printf("TTPs的私钥为：%s\n",SKs)
	fmt.Printf("TTPs的公钥为：%s\n",PKs)	
	//Data user注册密钥对，其中公钥pku公开在链上
	sku,pku:=ElGamal.EGSetup()
	fmt.Printf("Data user的私钥为：%s\n",sku)
	fmt.Printf("Data user的公钥为：%s\n",pku)
//-------------------Secret-Hiding-------------------//
    //随机生成一个明文信息
	m,_ := rand.Int(rand.Reader, order)
    //data owner加密明文信息,其中密文C公布在链上
	C:=Threshold_ElGamal.THEGEncrypt(m,pko)
	fmt.Printf("密文C为：%s\n",C)
	//生成私钥sko的PVSS份额Key，份额承诺Commitments以及{g^si},其中Commitments以及{g^si}公布在链上
	VSS_SK,Key:=Threshold_ElGamal.THEGKenGen(C, sko, numShares, threshold)
    fmt.Printf("映射后的秘密份额Gs为：%s\n", VSS_SK.Gs)
    fmt.Printf("秘密份额的承诺Commitments为：%s\n", VSS_SK.Commitments)
    fmt.Printf("私钥sko的PVSS份额Key为：%s\n", Key)
    
    //用TTPs的公钥加密Key得到CKey,并将CKey公布在链上
    CKey:=make([]*bn256.G1,numShares)
    for i:=0;i<numShares;i++{
    	CKey[i]=new(bn256.G1).Add(Key[i], new(bn256.G1).ScalarMult(PKs[i],VSS_SK.Shares[i]))
    }
    fmt.Printf("加密后的PVSS份额为：%s\n", CKey)
    ////生成DLEQ Proof prfs_s:(g,g^si,Key*pki^si,CKey,si),并将prfs_s公布在链上
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
    fmt.Printf("生成的prfs_s为：%s\n",prfs_s)  
    
    //生成DLEQ Proof prfs'_s:(g,g^si,C0,Key,si)
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
    fmt.Printf("生成的prfs'_s为：%s\n", _prfs_s)
//-------------------Key-Verification-------------------// 
	//VSS的验证结果（需要链上智能合约完成）
	result:=vss.VerifyShare(VSS_SK.Gs,VSS_SK.Commitments)
	fmt.Printf("VSS验证结果为：%v\n", result)
	//prfs_s的DLEQ验证结果（需要链上智能合约完成）
	Error:= dleq.Mul_Verify(prfs_s.C, prfs_s.Z, mul_G, mul_H, prfs_s.XG, prfs_s.XH, prfs_s.RG, prfs_s.RH)
    fmt.Printf("prfs_s的验证结果为：%v\n",Error)
//-------------------Key-Delegation-------------------// 
	//TTPs用data user的公钥对密钥进行加密得EKey，其中EKey被公布在链上  
	TTPs_Key:=make([]*bn256.G1,numShares) 
	for i:=0;i<numShares;i++{			TTPs_Key[i]=new(bn256.G1).Add(CKey[i],new(bn256.G1).Neg(new(bn256.G1).ScalarMult(VSS_SK.Gs[i],SKs[i])))
	} 
	EKey:=ElGamal.EGEncrypt(TTPs_Key,pku,numShares)
	fmt.Printf("TTPs加密后的密钥EKey为：%s\n",EKey)  
//-------------------Key-Delegation-------------------//
	_Key:= make([]*bn256.G1,numShares)
	for i:=0;i<numShares;i++{
		_Key=ElGamal.EGDecrypt(EKey,sku,numShares)
	}
	fmt.Printf("data user解密所得密钥_Key为：%s\n",_Key) 

	for i:=0;i<numShares;i++{
		_prfs_s.XH[i]=_Key[i]
    }
    //data user验证解密所得密钥_Key的正确性（由链上智能合约完成）
    Error= dleq.Mul_Verify(_prfs_s.C, _prfs_s.Z, mul_G, _mul_H, _prfs_s.XG, _prfs_s.XH, _prfs_s.RG, _prfs_s.RH)
    fmt.Printf("prfs'_s的验证结果为：%v\n",Error)
	//data user利用正确的密钥份额解密得到data owner的明文信息
    KeyIndices := make([]*big.Int, threshold)
	for i := 0; i < threshold; i++ {
		KeyIndices[i] = big.NewInt(int64(i + 1))
	}
	_m:=Threshold_ElGamal.THEGDecrypt(C, _Key, KeyIndices, threshold)
	fmt.Printf("data user解密后所得明文信息为：%s\n",_m)
//-------------------Dispute-------------------//
	//data user生成一个DIS,其中DIS公布在链上
	DIS:=new(bn256.G1).ScalarMult(EKey.EK0[0],sku)
	//data user生成sku的DLEQProof：prfs_sku，并将prfs_sku公布在链上
	_c, _z, _xG, _xH, _rG, _rH, _ := dleq.NewDLEQProof(g, EKey.EK0[0], sku)
	prfs_sku:=DLEQProof{C:_c, Z:_z, XG:_xG, XH:_xH, RG:_rG, RH:_rH} 
	fmt.Printf("生成的prfs_sku为：%s\n",prfs_sku)
	//验证纠纷是否正确（由链上智能合约完成）
 	_Error:=dleq.Verify(prfs_sku.C, prfs_sku.Z, g, EKey.EK0[0], pku, DIS, prfs_sku.RG, prfs_sku.RH)
	fmt.Printf("纠纷验证结果为：%v\n",_Error)
}
