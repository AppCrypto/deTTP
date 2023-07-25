from web3 import Web3
w3 = Web3(Web3.HTTPProvider('http://127.0.0.1:7545'))
from solcx import compile_standard,install_solc
install_solc("0.8.0")
import time
import json #to save the output in a JSON file
import time
#Compile and build the contract
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
#link test network
chain_id = 5777
accounts0 = w3.eth.accounts[0]
transaction_hash = contract.constructor().transact({'from': accounts0})
# deploy contract
transaction_receipt = w3.eth.wait_for_transaction_receipt(transaction_hash)
# Get the deployed contract address
contract_address = transaction_receipt['contractAddress']
print("约已部署，地址：", contract_address)

Contract = w3.eth.contract(address=contract_address, abi=abi)

accounts9 = w3.eth.accounts[9]

#Test gas for successful task execution
Contract.functions.new_task(accounts9,accounts0,30000000000000000000,3).transact({'from': accounts0})
for i in range(3):
	accounts = w3.eth.accounts[i]
	Contract.functions.TTP_EDA_i(1000000000000000000,10000000000000000000,20).transact({'from': accounts})
	Contract.functions.deposit(accounts0).transact({'from': accounts, 'value': 9000000000000000000})
	Contract.functions.record(accounts0,10).transact({'from': accounts})
fee=0
fee = Contract.functions.date_user_fee(accounts0).transact({'from': accounts0})

#根据每次任务更换value值，因为我没法做到将以太坊返回值fee变为下一行代码的value，所以采用了手动计算
Contract.functions.date_user_pay(accounts0).transact({'from': accounts0, 'value': 90000000000000000000 })

Contract.functions.success_distribute(accounts0).transact({'from': accounts0})
Contract.functions.updateCY_i(accounts0).transact({'from': accounts0})

#Test failed to execute the gas of the task
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
#There is a delay time that is less than the blockchain time set in the code, causing the code to report an error. 存在延迟时间小于代码中设置的测试区块链时间导致代码报错
time.sleep(120)
Contract.functions.fail_distribute(accounts4).transact({'from': accounts0})
Contract.functions.updateCY_i(accounts4).transact({'from': accounts0})


