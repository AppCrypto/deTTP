# Example of deployment smart contracts to EVM using GoLang

Here is a simple and convenient golang code utilizes `go-ethereum`,`abigen` and `solc` to deploy the smart contract and intract with the it.

Follow these steps to make it easy to deploy the contracts and intract with smart contracts.

# Pre requisites

* Ubuntu OS 

  Version: Ubuntu 22.04.4 LTS

* VS code

* Golang  
https://go.dev/dl/  
Version：1.22.0 linux/amd64

* Solidity  
https://docs.soliditylang.org/en/v0.8.2/installing-solidity.html  
    Version: 0.8.20

* Solidity compiler (solc).  
https://docs.soliditylang.org/en/latest/installing-solidity.html 

    Version: 0.8.25-develop

* Ganache-cli      
    ```bash
    npm install -g ganache  
    ```
    version：v7.9.2 (@ganache/cli: 0.10.2, @ganache/core: 0.10.2)

* Abigen   
    ```bash
    go get -u github.com/ethereum/go-ethereum
    cd $GOPATH/src/github.com/ethereum/go-ethereum/
    sudo make && make devtools 
    ```
    Version: v1.14.3

# Package description

* `Main.go`    main executable file, run this file to invoke the contract.

* `compile/contract/`  The folder for store solidity(.sol) contract file.

* `compile/compile.sh`  The script file that compile and generate ABI from a solidity for later use.

* `genPrvKey.sh`  The script file that run ganache for the first time to get the account key and generate the`.env` file.

* `.env`  The file to store private key for the default ganache account. 

* `utils/utils.go`  The file for deploying and compiling contracts.

* `crypto/`  The folder for store cryptographic primitives (EIGamal, Threshold_ElGamal,DLEQ, VSS).

# How to use it

