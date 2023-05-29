from typing import Dict, List, Tuple, Set, Optional
import sys
import time
import pprint
from solcx import set_solc_version,install_solc
import web3
from web3 import Web3
from solcx import compile_source
import utils
import random
# install_solc('v0.8.0')
set_solc_version('v0.8.0')

import sympy # consider removing this dependency, only needed for mod_inverse
import re
import numpy as np


from typing import Tuple, Dict, List, Iterable, Union
from py_ecc.typing import FQ,FQ2
from py_ecc.bn128 import G1, G2
from py_ecc.bn128 import add, multiply, neg, pairing, is_on_curve
from py_ecc.bn128 import curve_order as CURVE_ORDER
from py_ecc.bn128 import field_modulus as FIELD_MODULUS



import hashlib

def hash(s):
    x = hashlib.sha256()
    x.update(str(s).encode())
    return x.hexdigest()

def compile_source_file(file_path):
   with open(file_path, 'r') as f:
      source = f.read()
   return compile_source(source)


def deploy_contract(w3, contract_interface):
    # print(contract_interface)
    # accounts = web3.geth.personal.list_accounts()
    # if len(w3.eth.accounts) == 0:
    #     w3.geth.personal.new_account('123456')
    account=w3.eth.accounts[0]
    # w3.geth.personal.unlock_account(account,"123456")
    contract = w3.eth.contract(
        abi=contract_interface['abi'],
        bytecode=contract_interface['bin'])
    tx_hash = contract.constructor().transact({'from': account, 'gas': 500_000_000})

    # tx_hash = contract.constructor({'from': account, 'gas': 500_000_000}).transact()
    address = w3.eth.getTransactionReceipt(tx_hash)['contractAddress']
    return address


# 
w3=web3.Web3(web3.HTTPProvider('http://127.0.0.1:7550', request_kwargs={'timeout': 60 * 10}))

contract_source_path = './contracts/Verification.sol'
compiled_sol = compile_source_file(contract_source_path)
# print(compiled_sol)

# 
contract_id, contract_interface = compiled_sol.popitem()
address = deploy_contract(w3, contract_interface)
print("Deployed {0} to: {1}\n".format(contract_id, address))


ctt = w3.eth.contract(
   address=address,
   abi=contract_interface['abi'])
# print(contract_interface['abi'])




def random_scalar() -> int:
    """ Returns a random exponent for the BN128 curve, i.e. a random element from Zq.
    """
    return random.randint(0,CURVE_ORDER)


def share_secret(secret:int ,n:int ,t: int)-> Dict[int, int]:

    coefficients = [secret] + [random_scalar() for j in range(t-1)]
    #coefficients=[5]+[x for x in range(1,t)]
    
    #print(coefficients)
    def f(x: int) -> int:
        """ evaluation function for secret polynomial
        """
        return (
            sum(coef * pow(x, j, CURVE_ORDER) for j, coef in enumerate(coefficients)) % CURVE_ORDER
        )
    shares = { x:f(x) for x in range(1,n+1) }
    #print(shares)
    return shares

def vss_share_secret(secret:int ,n:int ,t: int):

    coefficients = [secret] + [random_scalar() for j in range(t-1)]
    #coefficients=[5]+[x for x in range(1,t)]
    
    #print(coefficients)
    def f(x: int) -> int:
        """ evaluation function for secret polynomial
        """
        return (
            sum(coef * pow(x, j, CURVE_ORDER) for j, coef in enumerate(coefficients)) % CURVE_ORDER
        )
    shares = { x:f(x) for x in range(1,n+1) }    
    gs = { i: multiply(G1, shares[i]) for i in range(1,n+1) }
    sgs = { i: shares[i]*int(hash(gs[i]),16) for i in range(1,n+1) }
    comj = {j: multiply(G1, coefficients[j]) for j in range(0,t) }
    
    #print(shares)
    return shares,gs,comj,sgs

def vss_verify(gs:Dict[int, int], comj:Dict[int, int]) -> bool:
   for i in gs:
      x=multiply(G1,0) 
      # print(type(i))
      for j in comj:
         x = add(x, multiply(comj[j], pow(i,j,CURVE_ORDER))) 
      # print(gs[i] ,x )
      if gs[i] != x :
         return False
   print("vss_verify",True)
   return True


def lagrange_coefficient(i: int,keys) -> int:
    result = 1
    for j in keys:
        if i != j:
            result *= j * sympy.mod_inverse((j - i) % CURVE_ORDER, CURVE_ORDER)
            result %= CURVE_ORDER
    # print(result)
    return result


def recover_secret(shares: Dict[int, int]) -> int:
    """ Recovers a shared secret from t VALID shares.
    """
    return sum(share * lagrange_coefficient(i, shares.keys()) for i, share in shares.items()) % CURVE_ORDER


def FQ2IntArr(fqArr):
   x=[]
   for fq in fqArr:
      x.append([fq[0].n, fq[1].n])
   return x

def FQ2IntArr2(fqArr):
   x=[]
   for fq in fqArr:
      x.append(fq[0].n)
      x.append(fq[1].n)
   return x


def dleq(g, y1, h, y2, shares):
     # print(len(g),len(y1),len(h),len(y2),len(shares))
     w = random_scalar()
     z=[0 for i in range(0,len(y1))]
     a1=[0 for i in range(0,len(y1))]
     a2=[0 for i in range(0,len(y1))]
     c = int(hash(str(y1)+str(y2)),16)
     
     for i in range(0, len(y1)):
         # print(i,i in y1,i in y2)
         a1[i] = multiply(g[i], w)
         if isinstance(h,list):
             a2[i] = multiply(h[i], w)
         else:
             a2[i] = multiply(h, w)
         z[i] = (w - shares[i+1] * c)  %  CURVE_ORDER 
     
     return c, a1, a2, z


def Dispute_dleq(g, y1, h, y2, SKu):
     # print(len(g),len(y1),len(h),len(y2),len(shares))
     w = random_scalar()
     a1=[0 for i in range(0,len(y2))]
     a2=[0 for i in range(0,len(y2))]
     c = int(hash(str(y1)+str(y2)),16)
     z = (w - SKu * c)  %  CURVE_ORDER
     for i in range(0, len(y2)):
         # print(i,i in y1,i in y2)
         a1[i] = multiply(g[i], w)
         a2[i] = multiply(h[i], w)
         a2[i] = multiply(h[i], w)
          
     
     return c, a1, a2, z

def dleq_verify(g, y1, h, y2, c, a1, a2, z):
    
     for i in range(0, len(g)):
         if isinstance(h,list):
             if a1[i] !=add(multiply(g[i], z[i]), multiply(y1[i+1], c)) \
             or a2[i] !=add(multiply(h[i], z[i]), multiply(y2[i+1], c)):
                 return False
         else:
             if a1[i] !=add(multiply(g[i], z[i]), multiply(y1[i+1], c)) \
             or a2[i] !=add(multiply(h, z[i]), multiply(y2[i+1], c)):
                 return False             
     # print("dleq_verify", True)
     return True
"""
def Dispute_dleq_verify(g, y1, h, y2, c, a1, a2, z):
     
     for i in range(0, len(g)):
         if a1[i] !=add(multiply(g[i], z), multiply(y1,c)) \
         or a2[i] !=add(multiply(h[i], z), multiply(y2[i], c)):
             return False
         
     # print("dleq_verify", True)
     return True
"""

def Dispute_dleq_verify(g, y1, h, y2, c, a1, a2, z):
    
     for i in range(0, len(g)):
         if isinstance(h,list):
             if a1[i] !=add(multiply(g[i], z[i]), multiply(y1[i], c)) \
             or a2[i] !=add(multiply(h[i], z[i]), multiply(y2[i], c)):
                 return False
         else:
             if a1[i] !=add(multiply(g[i], z[i]), multiply(y1[i], c)) \
             or a2[i] !=add(multiply(h, z[i]), multiply(y2[i], c)):
                 return False             
     # print("dleq_verify", True)
     return True

     
def Convert_type_list(data):
    data_list=[[int(x) for x in list] for list in data]
    return data_list

    


n=2
t=int(n/2)+1
# Registration
SKo=random_scalar()
PKo=multiply(G1, SKo)


SKu=random_scalar()


PKu=multiply(G1, SKu)



SKs=[0]
PKs=[multiply(G1, 0)]
for i in range(0, n):
   r=random_scalar()
   SKs.append(r)
   PKs.append(multiply(G1, r))

# # Data owner distribute
# THEGEncrypt
m=multiply(G1, random_scalar())
r=random_scalar()
C={"C0":multiply(G1, r), "C1":add(m, multiply(PKo,r))}
# starttime=time.time()
# for i in range(0,10):
#     C={"C0":multiply(G1, r), "C1":add(m, multiply(PKo,r))}
# print("elgamal encrypt cost:",(time.time()-starttime)/10. )


# THEGKeygen
starttime = time.time()
secret = SKo
shares,gs,comj,sgs=vss_share_secret(secret,n,t)
# vss_verify(gs,comj)
K={j: multiply(C["C0"], shares[j]) for j in shares}
CK={j: add(K[j], multiply(PKs[j],shares[j])) for j in shares}
# shares_for_recovery = dict(random.sample(shares.items(), t))
# print("test recover_secret",recover_secret(shares_for_recovery)==secret)

C={"C0":multiply(G1, r), "C1":add(m, multiply(PKo,r))}

c, a1, a2, z= dleq([G1 for i in shares], gs, [add(C['C0'],PKs[i]) for i in range(1, len(PKs))] ,CK,shares)

Kp_c, Kp_a1, Kp_a2, Kp_z =  dleq([G1 for i in shares], gs, [C['C0'] for i in range(1, len(PKs))], K, shares)

print("### Secret Hiding time cost", time.time() - starttime)

# Key Verification
starttime = time.time()
dleq_verify([G1 for i in shares], gs, [add(C['C0'],PKs[i]) for i in range(1, len(PKs))],CK,c, a1, a2, z)
print("dleq_verify time cost:",(time.time()- starttime)/len(PKs))

gsK=gs.keys()
gsV=[gs[k] for k in shares]
comjK=comj.keys()
comjV=[comj[k] for k in comj]


# VSSVerify
gas_estimate = ctt.functions.VSSVerify(list(gsK)+FQ2IntArr2(gsV)+list(comjK)+FQ2IntArr2(comjV),len(gsK),len(comjK)).estimateGas()
print("Sending transaction to VSSVerify ",gas_estimate)

ret = ctt.functions.VSSVerify(list(gsK)+FQ2IntArr2(gsV)+list(comjK)+FQ2IntArr2(comjV),len(gsK),len(comjK)).call({'from':w3.eth.accounts[0],'gas': 500_000_000})
# print("Sending transaction to VSSVerify ",ret)
assert(ret)


g=Convert_type_list([G1 for i in shares])
y1=Convert_type_list(gs.values())
h=Convert_type_list([add(C['C0'],PKs[i]) for i in range(1, len(PKs))])
y2=Convert_type_list(CK.values())


#DLEQ_SmartContract
gas_estimate_DELQ=ctt.functions.DELQVerify(g,y1,h,y2,c,Convert_type_list(a1),Convert_type_list(a2),z).estimateGas()
print("Sending transaction to DELQVerify ",gas_estimate_DELQ)
ret_DELQ = ctt.functions.DELQVerify(g,y1,h,y2,c,Convert_type_list(a1),Convert_type_list(a2),z).call({'from':w3.eth.accounts[0],'gas': 500_000_000})
# print("Sending transaction to DELQVerify ",ret_DELQ)
assert(ret)

# TODO data owner uploads CK to smart contract(finish)
ctt.functions.UploadCK(Convert_type_list(CK.values())).transact({'from':w3.eth.accounts[0],'gas': 500_000_000})
# print("Sending transaction to UploadCK")
gas_estimate_UploadCK=ctt.functions.UploadCK(Convert_type_list(CK.values())).estimateGas()
print("The gas of uploading CK ",gas_estimate_UploadCK)

# print(K)
# Key Delegation

# TODO TTP downloads CK from smart contract(finish)
ctt.functions.DownloadCK().transact({'from':w3.eth.accounts[0],'gas': 500_000_000})
# print("Sending transaction to DownloadloadCK")

gas_estimate_DownloadCK=ctt.functions.DownloadCK().estimateGas()
print("The gas of downloading CK ",gas_estimate_DownloadCK)

starttime = time.time()
Kp={j: add(CK[j],neg(multiply(gs[j],SKs[j]))) for j in CK} #TTP extracts K from CK
assert(Kp==K)

EK={}
for j in K:
    l=random_scalar()
    EK[j]={"EK0":multiply(G1, l), "EK1":add(K[j], multiply(PKu,l))}#hide K into EK

print("### Key Delegation time cost", time.time() - starttime)

EK_0=[]
EK_1=[]

for k in EK:
    EK_0.append(EK[j].get("EK0"))
    EK_1.append(EK[j].get("EK1"))

# TODO TTP uploads EK to smart contract(finish)
ctt.functions.UploadEK(Convert_type_list(EK_0),Convert_type_list(EK_1)).transact({'from':w3.eth.accounts[0],'gas': 500_000_000})
print("Sending transaction to UploadEK")
gas_estimate_UploadEK=ctt.functions.UploadEK(Convert_type_list(EK_0),Convert_type_list(EK_1)).estimateGas()
print("The gas of uploading EK ",gas_estimate_UploadEK)

# TODO data user downloads EK from smart contract(finish)
EK_contract_0=ctt.functions.DownloadEK_0().transact({'from':w3.eth.accounts[0],'gas': 500_000_000})
# print("Sending transaction to DownloadloadEK_0")
gas_estimate_DownloadEK_0=ctt.functions.DownloadEK_0().estimateGas()
print("The gas of downloading EK_0 ",gas_estimate_DownloadEK_0)

EK_contract_1=ctt.functions.DownloadEK_1().transact({'from':w3.eth.accounts[0],'gas': 500_000_000})
# print("Sending transaction to DownloadloadEK_1")
gas_estimate_DownloadEK_1=ctt.functions.DownloadEK_1().estimateGas()
print("The gas of downloading EK_1 ",gas_estimate_DownloadEK_1)
starttime=time.time()
Kp={j: add(EK[j]["EK1"],neg(multiply(EK[j]["EK0"],SKu))) for j in EK} #Data user extracts K from EK
print("elgamal decrypt cost:",(time.time()-starttime)/len(EK) )

assert(Kp==K)
# TODO test equation 5 off the blockchain(finish)
_c, _a1, _a2, _z= dleq([G1 for i in shares], gs, C['C0'] ,K,shares)
DELQ_result=dleq_verify([G1 for i in shares], gs, C['C0'], K , _c, _a1, _a2, _z)
assert(DELQ_result)

# TODO upload dispute to smart contract, i.e., equation 6(finish)
Dis=[]  
for j in EK_0:
    Dis.append(multiply(j,SKu))

Dis_c, Dis_a1, Dis_a2, Dis_z= Dispute_dleq([G1 for i in shares], PKu, EK_0 ,Dis,SKu)
result=Dispute_dleq_verify([G1 for i in EK_0], [PKu for i in EK_0], EK_0,Dis, Dis_c, Dis_a1, Dis_a2, [Dis_z for i in EK_0])
print("result=",result)

ctt.functions.UploadDispute(Convert_type_list([G1 for i in EK_0]), Convert_type_list([PKu for i in EK_0]), Convert_type_list(EK_0) ,Convert_type_list(Dis),Dis_c ,Convert_type_list(Dis_a1),Convert_type_list(Dis_a2),[Dis_z for i in EK_0]).transact({'from':w3.eth.accounts[0],'gas': 500_000_000})
print("Sending transaction to UploadDispute")
gas_estimate_UploadDispute=ctt.functions.UploadDispute(Convert_type_list([G1 for i in EK_0]), Convert_type_list([PKu for i in EK_0]), Convert_type_list(EK_0) ,Convert_type_list(Dis),Dis_c ,Convert_type_list(Dis_a1),Convert_type_list(Dis_a2),[Dis_z for i in EK_0]).estimateGas()
print("The gas of uploading dispute ",gas_estimate_UploadDispute)


# TODO Verify dispute


gas_estimate_Dis_DELQ=ctt.functions.DELQVerify(Convert_type_list([G1 for i in EK_0]), Convert_type_list([PKu for i in EK_0]), Convert_type_list(EK_0) ,Convert_type_list(Dis),Dis_c ,Convert_type_list(Dis_a1),Convert_type_list(Dis_a2),[Dis_z for i in EK_0]).estimateGas()
print("The gas of dispute DELQVerify is",gas_estimate_Dis_DELQ)
ret_Dis_DELQ = ctt.functions.DELQVerify(Convert_type_list([G1 for i in EK_0]), Convert_type_list([PKu for i in EK_0]), Convert_type_list(EK_0) ,Convert_type_list(Dis),Dis_c ,Convert_type_list(Dis_a1),Convert_type_list(Dis_a2),[Dis_z for i in EK_0]).call({'from':w3.eth.accounts[0],'gas': 500_000_000})
print("The result of dispute DELQVerify is",ret_Dis_DELQ)



# EGDecrypt
starttime = time.time()
Kp={j: add(EK[j]["EK1"],neg(multiply(EK[j]["EK0"],SKu))) for j in EK} #Data user extracts K from EK
# Verify Kp
dleq_verify([G1 for i in shares], gs, [C['C0'] for i in range(1, len(PKs))],Kp,Kp_c, Kp_a1, Kp_a2, Kp_z)
# THEGDecrypt
W=multiply(G1, 0)
for i in Kp:    
    tmp = multiply(Kp[i], lagrange_coefficient(i, Kp.keys()))
    W = add(W, tmp)
mp=add(C['C1'],neg(W))
print("### Secret Recovery time cost", time.time() - starttime)

print("data user obtain data owner's secret", mp==m)    

