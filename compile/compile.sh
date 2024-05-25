solc --evm-version paris --optimize --abi ./contract/Verification.sol -o contract --overwrite
solc --evm-version paris --optimize --bin ./contract/Verification.sol -o contract --overwrite
abigen --abi=./contract/Verification.abi --bin=./contract/Verification.bin --pkg=contract --out=./contract/Verification.go