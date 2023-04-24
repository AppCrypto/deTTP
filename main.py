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


import hashlib

def hash(str):
    x = hashlib.sha256()
    x.update(str.encode())
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

import pysolcrypto.schnorr



s = int(random.random()*(2**256))#19977808579986318922850133509558564821349392755821541651519240729619349670944
m = int(hash("msg to be verified"),16)#19996069338995852671689530047675557654938145690856663988250996769054266469975

guankong1=w3.eth.accounts[1]
guankong2=w3.eth.accounts[2]
guankong3=w3.eth.accounts[3]

putong1=w3.eth.accounts[4]#与guankong1进行消息沟通
putong2=w3.eth.accounts[5]#与guankong1进行消息沟通
putong3=w3.eth.accounts[6]#与guankong2进行消息沟通
putong4=w3.eth.accounts[7]#与guankong2进行消息沟通
putong5=w3.eth.accounts[8]#与guankong3进行消息沟通
putong6=w3.eth.accounts[9]#与guankong3进行消息沟通
import time

for i in range(1, 10):   
   gas_estimate = ctt.functions.register("id"+str(i)).estimateGas()
   print("Sending transaction to register ",gas_estimate)
   ret = ctt.functions.register("id"+str(i)).transact({"from":w3.eth.accounts[0], 'gas': 500_000_000})
   # print("register successful:",ret)

starttime=time.time()
senderID="id1"

proof = list(pysolcrypto.schnorr.schnorr_create(s, m, senderID))
for i in range(0, len(proof[0])):
   proof[0][i]=proof[0][i].n

gas_estimate = ctt.functions.VerifySchnorrProof(proof[0],m+int(hash(senderID)[:16],16),proof[1],proof[2],senderID).estimateGas()
print("Sending transaction to VerifySchnorrProof ",gas_estimate)
ret = ctt.functions.VerifySchnorrProof(proof[0],m+int(hash(senderID)[:16],16),proof[1],proof[2],senderID).call({'from':w3.eth.accounts[0],'gas': 500_000_000})
print("Schnorr verify:",ret)
print("verification time cost",time.time()-starttime)

senderID="id2"
ret = ctt.functions.VerifySchnorrProof(proof[0],m+int(hash(senderID)[:16],16),proof[1],proof[2],senderID).call({'from':w3.eth.accounts[0],'gas': 500_000_000})
print("Schnorr verify:",ret,", when sender ID is changed")







# exit()