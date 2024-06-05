package main

import (
	"context"
	"dttp/compile/contract"
	"dttp/utils"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	//Number of ttp nodes (begin 0 )
	n := 10

	contract_name := "Verification"

	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	owner_Address := common.HexToAddress(utils.GetENV("ACCOUNT_2"))
	user_Address := common.HexToAddress(utils.GetENV("ACCOUNT_3"))
	TTP_Address := common.HexToAddress(utils.GetENV("ACCOUNT_4"))
	fTTP_Address := common.HexToAddress(utils.GetENV("ACCOUNT_5"))
	privatekey1 := utils.GetENV("PRIVATE_KEY_1")

	auth1 := utils.Transact(client, privatekey1, big.NewInt(0))

	address, _ := utils.Deploy(client, contract_name, auth1)

	Contract, err := contract.NewContract(common.HexToAddress(address.Hex()), client)
	if err != nil {
		fmt.Println(err)
	}

	//--------------------------------Successful task
	auth2 := utils.Transact(client, privatekey1, big.NewInt(0))
	//Create a successful settlement task
	Contract.NewTask(auth2, owner_Address, user_Address, big.NewInt(1000000000000000000), big.NewInt(int64(n)))
	//TTP Registration
	for i := 0; i < n; i++ {

		auth3 := utils.Transact(client, privatekey1, big.NewInt(0))

		var (
			CV_i = big.NewInt(100000000000000000)
			EV_i = big.NewInt(1000000000000000000)
			RP_i = big.NewInt(20)
		)

		tx1, err := Contract.TTPRegister(auth3, CV_i, EV_i, RP_i, TTP_Address)
		if err != nil {
			fmt.Println(err)
		}
		receipt1, err := bind.WaitMined(context.Background(), client, tx1)
		if err != nil {
			log.Fatalf("Tx receipt failed: %v", err)
		}
		fmt.Printf("TTP%d registration and calculation of EDAI Gas used: %d\n", i, receipt1.GasUsed)
	}

	for i := 0; i < n; i++ {
		//Query TTP deposited funds
		_, _, _, EDA_i, _, _ := Contract.QueryTTP(&bind.CallOpts{}, big.NewInt(int64(i)))

		privatekey4 := utils.GetENV("PRIVATE_KEY_4")

		auth4 := utils.Transact(client, privatekey4, EDA_i)
		//Deposit money
		tx2, err := Contract.Deposit(auth4, big.NewInt(int64(i)), big.NewInt(0))
		if err != nil {
			fmt.Println(err)
		}
		receipt2, err := bind.WaitMined(context.Background(), client, tx2)
		if err != nil {
			log.Fatalf("Tx receipt failed: %v", err)
		}
		fmt.Printf("TTP%d deposit money Gas used: %d\n", i, receipt2.GasUsed)
	}

	auth5 := utils.Transact(client, privatekey1, big.NewInt(0))
	//Calculate the amount of money data_user needs to pay
	_, err = Contract.DateUserFee(auth5, big.NewInt(0))
	if err != nil {
		fmt.Println(err)
	}
	//Query the amount of money data_user needs to pay
	DateUserFee, err := Contract.QueryDateUserFee(&bind.CallOpts{}, big.NewInt(0))
	if err != nil {
		fmt.Println(err)
	}

	privatekey3 := utils.GetENV("PRIVATE_KEY_3")

	auth6 := utils.Transact(client, privatekey3, DateUserFee)
	//data_user pays
	_, err = Contract.DateUserPay(auth6, big.NewInt(0))
	if err != nil {
		fmt.Println(err)
	}

	//Create a queue provided by vss that has passed verification
	success := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		success[i] = big.NewInt(int64(i))
	}

	auth7 := utils.Transact(client, privatekey1, big.NewInt(0))
	//Calling a successful task
	tx3, err := Contract.SuccessDistribute(auth7, big.NewInt(0), success)
	if err != nil {
		fmt.Println(err)
	}
	receipt3, err := bind.WaitMined(context.Background(), client, tx3)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}

	fmt.Printf("successful task Gas used: %d\n", receipt3.GasUsed)

	auth8 := utils.Transact(client, privatekey1, big.NewInt(0))
	//update cv_i
	tx4, err := Contract.UpdateCYI(auth8, big.NewInt(0), success)
	if err != nil {
		fmt.Println(err)
	}
	receipt4, err := bind.WaitMined(context.Background(), client, tx4)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}

	fmt.Printf("updata cv_i: %d\n", receipt4.GasUsed)

	// Get address balance
	balance1, err := client.BalanceAt(context.Background(), owner_Address, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("owner_Address %s balance: %d Wei\n", owner_Address, balance1)

	balance2, err := client.BalanceAt(context.Background(), user_Address, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("user_Address %s balance: %d Wei\n", user_Address.Hex(), balance2)

	balance3, err := client.BalanceAt(context.Background(), TTP_Address, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("TTP_Address %s balance: %d Wei\n", TTP_Address.Hex(), balance3)
	//--------------------------------Failed task
	auth22 := utils.Transact(client, privatekey1, big.NewInt(0))
	Contract.NewTask(auth22, owner_Address, user_Address, big.NewInt(1000000000000000000), big.NewInt(int64(n)))
	for i := 0; i < (n - 1); i++ {
		//Query TTP deposited funds
		_, _, _, EDA_i, _, _ := Contract.QueryTTP(&bind.CallOpts{}, big.NewInt(int64(i)))

		privatekey4 := utils.GetENV("PRIVATE_KEY_4")
		//Deposit money
		auth9 := utils.Transact(client, privatekey4, EDA_i)
		tx2, err := Contract.Deposit(auth9, big.NewInt(int64(i)), big.NewInt(1))
		if err != nil {
			fmt.Println(err)
		}

		receipt7, err := bind.WaitMined(context.Background(), client, tx2)
		if err != nil {
			log.Fatalf("Tx receipt failed: %v", err)
		}
		fmt.Printf("TTP%d deposit money Gas used: %d\n", i, receipt7.GasUsed)

	}

	////TTP pledge funds for failed verification (in order to see the penalty on the account)
	_, _, _, fEDA_i, _, _ := Contract.QueryTTP(&bind.CallOpts{}, big.NewInt(int64(n-1)))

	fprivatekey4 := utils.GetENV("PRIVATE_KEY_5")

	fauth4 := utils.Transact(client, fprivatekey4, fEDA_i)

	_, err = Contract.Deposit(fauth4, big.NewInt(int64(n-1)), big.NewInt(1))
	if err != nil {
		fmt.Println(err)
	}

	auth10 := utils.Transact(client, privatekey1, big.NewInt(0))
	//Calculate the amount of money data_user needs to pay
	_, err = Contract.DateUserFee(auth10, big.NewInt(1))
	if err != nil {
		fmt.Println(err)
	}

	//Query the amount of money data_user needs to pay
	DateUserFee, err = Contract.QueryDateUserFee(&bind.CallOpts{}, big.NewInt(1))
	if err != nil {
		fmt.Println(err)
	}

	privatekey3 = utils.GetENV("PRIVATE_KEY_3")

	auth11 := utils.Transact(client, privatekey3, DateUserFee)
	//data_user pays
	_, err = Contract.DateUserPay(auth11, big.NewInt(1))
	if err != nil {
		fmt.Println(err)
	}

	//Create a queue provided by vss that has passed verification
	success1 := make([]*big.Int, n-1)
	for i := 0; i < (n - 1); i++ {
		success1[i] = big.NewInt(int64(i))
	}

	auth12 := utils.Transact(client, privatekey1, big.NewInt(0))
	//Calling a failed task
	tx6, err := Contract.FailDistribute(auth12, big.NewInt(1), success1)
	if err != nil {
		fmt.Println(err)
	}
	receipt6, err := bind.WaitMined(context.Background(), client, tx6)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}

	fmt.Printf("failed task Gas used: %d\n", receipt6.GasUsed)
	auth13 := utils.Transact(client, privatekey1, big.NewInt(0))
	//update cv_i
	_, err = Contract.UpdateCYI(auth13, big.NewInt(1), success)
	if err != nil {
		fmt.Println(err)
	}

	// Get address balance
	fbalance1, err := client.BalanceAt(context.Background(), owner_Address, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("owner_Address %s balance: %d Wei\n", owner_Address, fbalance1)
	fbalance2, err := client.BalanceAt(context.Background(), user_Address, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("user_Address %s balance: %d Wei\n", user_Address.Hex(), fbalance2)
	fbalance3, err := client.BalanceAt(context.Background(), TTP_Address, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("TTP_Address %s balance: %d Wei\n", TTP_Address.Hex(), fbalance3)

	fbalance4, err := client.BalanceAt(context.Background(), fTTP_Address, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("filed_TTP_Address %s balance: %d Wei\n", fTTP_Address.Hex(), fbalance4)

}
