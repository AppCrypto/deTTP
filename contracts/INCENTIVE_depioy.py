
from web3 import Web3
w3 = Web3(Web3.HTTPProvider('http://127.0.0.1:7545'))
from solcx import compile_standard,install_solc
install_solc("0.8.0")
import time
import json #to save the output in a JSON file
import time

with open("/home/ma/Documents/code/solidity/new/INCENTIVE4.sol", "r") as file:
    contact_list_file = file.read()

compiled_sol = compile_standard(
    {
        "language": "Solidity",
        "sources": {"INCENTIVE4.sol": {"content": contact_list_file}},
        "settings": {
            "outputSelection": {
                "*": {
                     "*": ["abi", "metadata", "evm.bytecode", "evm.bytecode.sourceMap"] # output needed to interact with and deploy contract 
                }
            }
        },
    },
    solc_version="0.8.0",
)

#print(compiled_sol)
with open("compiled_code.json", "w") as file:
    json.dump(compiled_sol, file)
# get bytecode
bytecode = compiled_sol["contracts"]["INCENTIVE4.sol"]["INCENTIVE4"]["evm"]["bytecode"]["object"]
# get abi
abi = json.loads(compiled_sol["contracts"]["INCENTIVE4.sol"]["INCENTIVE4"]["metadata"])["output"]["abi"]
# Create the contract in Python
contract = w3.eth.contract(abi=abi, bytecode=bytecode)



chain_id = 5777
accounts0 = w3.eth.accounts[0]
transaction_hash = contract.constructor().transact({'from': accounts0})
# 等待合约部署完成
transaction_receipt = w3.eth.wait_for_transaction_receipt(transaction_hash)
# 获取部署后的合约地址
contract_address = transaction_receipt['contractAddress']
print("约已部署，地址：", contract_address)

Contract = w3.eth.contract(address=contract_address, abi=abi)

accounts9 = w3.eth.accounts[9]

Contract.functions.new_task(accounts9,accounts0,30000000000000000000,3).transact({'from': accounts0})
for i in range(3):
	accounts = w3.eth.accounts[i]
	Contract.functions.TTP_EDA_i(1000000000000000000,10000000000000000000,20).transact({'from': accounts})
	Contract.functions.deposit(accounts0).transact({'from': accounts, 'value': 9000000000000000000})
	Contract.functions.record(accounts0,10).transact({'from': accounts})
fee=0
fee = Contract.functions.date_user_fee(accounts0).transact({'from': accounts0})

Contract.functions.date_user_pay(accounts0).transact({'from': accounts0, 'value': 90000000000000000000 })

Contract.functions.success_distribute(accounts0).transact({'from': accounts0})
Contract.functions.updateCY_i(accounts0).transact({'from': accounts0})

accounts4 = w3.eth.accounts[4]
Contract.functions.new_task(accounts9,accounts4,30000000000000000000,4).transact({'from': accounts4})
for i in range(3):
	accounts = w3.eth.accounts[i]
	Contract.functions.TTP_EDA_i(1000000000000000000,10000000000000000000,20).transact({'from': accounts})
	Contract.functions.deposit(accounts4).transact({'from': accounts, 'value': 9000000000000000000})
	Contract.functions.record(accounts4,20).transact({'from': accounts})
accounts8 = w3.eth.accounts[8]
Contract.functions.TTP_EDA_i(1000000000000000000,10000000000000000000,20).transact({'from': accounts8})
Contract.functions.deposit(accounts4).transact({'from': accounts8, 'value': 9000000000000000000})
Contract.functions.record(accounts4,10).transact({'from': accounts8})
time.sleep(360)
Contract.functions.fail_distribute(accounts4).transact({'from': accounts0})
Contract.functions.updateCY_i(accounts4).transact({'from': accounts0})


