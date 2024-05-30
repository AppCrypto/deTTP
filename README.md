# Example of deployment smart contracts to EVM using GoLang

Here is a simple and convenient golang code utilizes `go-ethereum`,`abigen` and `solc` to deploy the smart contract and intract with the smart contract.

Follow these steps to make it easy to deploy the contracts and intract with smart contracts.

# Pre requisites

* `Ubuntu OS`  Version: Ubuntu 22.04.4 LTS

* `VS code`

* `Golang`  https://go.dev/dl/   Version：1.22.0 linux/amd64

* `Solidity`  https://docs.soliditylang.org/en/v0.8.2/installing-solidity.html  Version: 0.8.20

* `Solidity compiler (solc)`  https://docs.soliditylang.org/en/latest/installing-solidity.html  
Version: 0.8.25-develop

* `Ganache-cli`    Version：v7.9.2 (@ganache/cli: 0.10.2, @ganache/core: 0.10.2)

    ```bash
    npm install -g ganache  
    ```
    
* `Abigen`    Version: v1.14.3
    ```bash
    go get -u github.com/ethereum/go-ethereum
    cd $GOPATH/src/github.com/ethereum/go-ethereum/
    sudo make && make devtools 
    ```
    

# Package description

* `Main.go`    Main executable file, run this file to invoke the contract.

* `compile/contract/`  The folder for store solidity(.sol) contract file.

* `compile/compile.sh`  The script file that compile and generate ABI from a solidity for later use.

* `genPrvKey.sh`  The script file that run ganache for the first time to get the account key and generate the`.env` file.

* `.env`  The file to store private key for the default ganache account. 

* `utils/utils.go`  The file for deploying and compiling contracts.

* `crypto/`  The folder for store cryptographic primitives (EIGamal, Threshold_ElGamal,DLEQ, VSS).

# How to use it

1. Before compiling and deploying the smart contract, we initialize ganache by running `genPrvKey.sh` to get the private keys for the default accounts (the default is 10) and store them in the `.env` file.

    Next time start ganacha, just run as following: 

    ```bash
    ganache --mnemonic "dttp"
    ```

2. We put the contract file (`.sol`) in the `compile/contract/` folder, in this case `SimpleStorage.sol`.

3. Then we change the Name in `compile.sh` to be the Name of the contract, in this case `Name=SimpleStorage`, and run `compile.sh` to generate `.abi` and `.bin` and `.go` files.

    ```bash
    bash compile.sh
    ```

4. Last, we change the contract_name in `main.go` to be the name of the contract, in this case `contract_name := "SimpleStorage"`. And we run the code as following:
    ```bash
    go run main.go
    ```
    Note that each file has detailed comments.
