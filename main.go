package main

//激励机制成功结算
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
	//ttp节点数
	n := 6

	contract_name := "Verification"

	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	owner_Address := common.HexToAddress(utils.GetENV("ACCOUNT_2"))
	user_Address := common.HexToAddress(utils.GetENV("ACCOUNT_3"))
	TTP_Address := common.HexToAddress(utils.GetENV("ACCOUNT_4"))

	privatekey1 := utils.GetENV("PRIVATE_KEY_1")

	auth1 := utils.Transact(client, privatekey1, big.NewInt(0))

	address, _ := utils.Deploy(client, contract_name, auth1)

	Contract, err := contract.NewContract(common.HexToAddress(address.Hex()), client)
	if err != nil {
		fmt.Println(err)
	}

	auth2 := utils.Transact(client, privatekey1, big.NewInt(0))
	//1、创建任务
	Contract.NewTask(auth2, owner_Address, user_Address, big.NewInt(1000000000000000000), big.NewInt(int64(n)))

	fmt.Printf("创建新任务-----------------------------------\n")

	for i := 0; i < n; i++ {

		auth3 := utils.Transact(client, privatekey1, big.NewInt(0))

		var (
			CV_i = big.NewInt(100000000000000000)
			EV_i = big.NewInt(1000000000000000000)
			RP_i = big.NewInt(20)
		)
		//2、TTP注册
		tx1, err := Contract.TTPRegister(auth3, CV_i, EV_i, RP_i, TTP_Address)
		if err != nil {
			fmt.Println(err)
		}

		receipt1, err := bind.WaitMined(context.Background(), client, tx1)
		if err != nil {
			log.Fatalf("Tx receipt failed: %v", err)
		}
		fmt.Printf("注册TTP%d计算EVI Gas used: %d\n", i, receipt1.GasUsed)
		fmt.Printf("TTP%d注册-----------------------------------\n", i)
		//3、查询TTP质押资金
		_, _, _, EDA_i, _, _ := Contract.QueryTTP(&bind.CallOpts{}, big.NewInt(int64(i)))

		privatekey4 := utils.GetENV("PRIVATE_KEY_4")

		auth4 := utils.Transact(client, privatekey4, EDA_i)
		//4、质押资金
		tx2, err := Contract.Deposit(auth4, big.NewInt(int64(i)), big.NewInt(0))
		if err != nil {
			fmt.Println(err)
		}
		receipt2, err := bind.WaitMined(context.Background(), client, tx2)
		if err != nil {
			log.Fatalf("Tx receipt failed: %v", err)
		}
		fmt.Printf("TTP%d质押资金 Gas used: %d\n", i, receipt2.GasUsed)
		fmt.Printf("TTP%d质押-----------------------------------\n", i)
	}

	auth5 := utils.Transact(client, privatekey1, big.NewInt(0))
	//5、计算data_user需要支付的资金
	_, err = Contract.DateUserFee(auth5, big.NewInt(0))
	if err != nil {
		fmt.Println(err)
	}
	//6、查询data_user需要支付的资金
	DateUserFee, err := Contract.QueryDateUserFee(&bind.CallOpts{}, big.NewInt(0))
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("data_user_fee:%s-----------------------------------\n", DateUserFee)

	privatekey3 := utils.GetENV("PRIVATE_KEY_3")

	auth6 := utils.Transact(client, privatekey3, DateUserFee)
	//7、data_user付钱
	_, err = Contract.DateUserPay(auth6, big.NewInt(0))
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("data_user 已付款-----------------------------------\n")
	//创建验证通过的vss提供的数据
	success := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		success[i] = big.NewInt(int64(i))
	}

	auth7 := utils.Transact(client, privatekey1, big.NewInt(0))
	//8、调用成功验证的函数（结算任务）
	tx3, err := Contract.SuccessDistribute(auth7, big.NewInt(0), success)
	if err != nil {
		fmt.Println(err)
	}
	receipt3, err := bind.WaitMined(context.Background(), client, tx3)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}

	fmt.Printf("成功执行的 Gas used: %d\n", receipt3.GasUsed)

	auth8 := utils.Transact(client, privatekey1, big.NewInt(0))
	//9、更新信誉值
	tx4, err := Contract.UpdateCYI(auth8, big.NewInt(0), success)
	if err != nil {
		fmt.Println(err)
	}
	receipt4, err := bind.WaitMined(context.Background(), client, tx4)
	if err != nil {
		log.Fatalf("Tx receipt failed: %v", err)
	}

	fmt.Printf("更新cv_i: %d\n", receipt4.GasUsed)

	// 获取地址余额
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

}
