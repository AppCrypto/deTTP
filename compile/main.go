package main

import "dttp/utils"

func main() {
	//合约名称-------------------------------------------------(修改处)
	contract_name := "SimpleStorage"
	utils.Compile(contract_name)
	//将合约变为go语言接口
	utils.Abigen(contract_name)
}
