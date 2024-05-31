// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;


// https://www.iacr.org/cryptodb/archive/2002/ASIACRYPT/50/50.pdf
contract Verification
{
  	//存储TTP信息
    struct TTP {
        int256 CV_i;      //credit value
        int256 EV_i;      //expected valu
        int256 RP_i;      //Refundable percentage*100
        int256 EDA_i;     //expected digital assets
		address account;
    }
    struct task {
        uint  tasktime;
        address date_owner;     
        address date_user; 
        uint date_fee;    
        address[] TTPS;    
        uint256 n;       //The number of ttp required to complete the task
 		//Complete the ttp of the deposit data.        
		uint[] sendersdata;
        //Store verified address
        uint[] senderss;
        //Store unverified addressesuint
        uint[] sendersf;
        //Complete the ttp of the deposit funds.
        uint[] sendersa;
        int success_distribute1;
        int fail_distribute1;
		uint256 ALL_fee;
    }


    TTP[] TTPS;
    task[] tasks;

    

    int public MDA_i=50;  //minimum deposited assets
    int public a=6;
    int public b=3;


    // A mapping to store the ether balance of each user
    mapping(uint => mapping(uint => uint)) public balances;

		function new_task(address date_owner, address date_user, uint date_fee, uint256 n) public   returns (uint)
    {
         // 初始化一个新的 Task 对象
        task memory newTask;
        newTask.tasktime = block.timestamp;
        newTask.date_owner = date_owner;
        newTask.date_user = date_user;
        newTask.date_fee = date_fee;
        newTask.n = n;

        // 初始化其他字段，使用默认值或空值
        newTask.TTPS = new address[](0);
        newTask.sendersdata = new uint[](0);
        newTask.senderss = new uint[](0);
        newTask.sendersf = new uint[](0);
        newTask.sendersa = new uint[](0);
        newTask.success_distribute1 = 0;
        newTask.fail_distribute1 = 0;
        newTask.ALL_fee = 0;
        tasks.push(newTask);
        return tasks.length - 1;
    }

    //Function to calculate EDA_i
    function TTP_register(int256 CV_i, int256 EV_i, int256 RP_i, address account) public  returns (uint,int){
       
        int EDA_i;
        int A;

        A=a * EV_i * RP_i / 100 - b * CV_i;
        if (A >  MDA_i) {
            EDA_i = A;
        }
       
        else {
            EDA_i= MDA_i;
        }
        uint id = TTPS.length;
        TTPS.push(TTP(CV_i,EV_i,RP_i,EDA_i,account));
        return (id,EDA_i);
    }
    //Query TTP information
    function query_TTP(uint id) public view returns (int256, int256, int256, int256,address) {
        return (TTPS[id].CV_i, TTPS[id].EV_i, TTPS[id].RP_i, TTPS[id].EDA_i,TTPS[id].account);
    }  

    // A function to deposit ether to the contract
    function deposit(uint TTP_id,uint task_id) public payable {
        TTP memory ttp = TTPS[TTP_id];
        int256 A= ttp.EDA_i;
        uint256 B;
        B = balances[TTP_id][task_id];  
        require( B == 0, "You have already deposited");              
        require(msg.value == uint256(A), "You must send  EDA_i wei");
        balances[TTP_id][task_id] = msg.value;
        tasks[task_id].sendersa.push(TTP_id);
    }	

    //date_user pay
    function date_user_fee(uint task_id) public returns (uint256) { 
       uint256 ALL_fees=0;
       for (uint i = 0; i < tasks[task_id].sendersa.length; i++) {
            uint TTP_id = tasks[task_id].sendersa[i];
            ALL_fees += uint(TTPS[TTP_id].EV_i);
        }   
       tasks[task_id].ALL_fee = ALL_fees + tasks[task_id].date_fee;
       return (tasks[task_id].ALL_fee);
    }

    function query_date_user_fee(uint task_id) public view returns (uint256) {
        return tasks[task_id].ALL_fee;
    }  
    
      //date_user pay
    function date_user_pay(uint task_id) public payable {
       require(tasks[task_id].ALL_fee == msg.value, "The amount you sent is wrong");
    }	  

    //Allocation of Funds for Successful Mission Execution
    function success_distribute(uint task_id,uint[] memory success) public  {
        require(block.timestamp <= tasks[task_id].tasktime +  2 minutes, "Not enough time passed");
        require(success.length >= tasks[task_id].n, "The number of ttp has not reached the threshold");  
        uint[] memory temp = new uint[](tasks[task_id].sendersa.length-success.length);
		uint count = 0;
		for (uint i = 0; i < tasks[task_id].sendersa.length; i++) {
            bool found = false;
            for (uint j = 0; j < success.length; j++) {
                if (tasks[task_id].sendersa[i] == success[j]) {
                    found = true;
                    break;
                }
            }
            if (!found) {
				temp[count] = tasks[task_id].sendersa[i];
                count++;
            }
        }		
		//给验证失败的TTP发钱+质押后不提供数据的退钱
        for (uint i = 0; i < temp.length; i++) {
			TTP memory ttp = TTPS[temp[i]];
            address payable recipient1 = payable(ttp.account);
            uint refund = balances[temp[i]][task_id] * uint(ttp.RP_i) / 100;
            recipient1.transfer(refund);
            balances[temp[i]][task_id] -= refund;
        }
        uint ALL=0;
        for (uint i = 0; i < temp.length; i++) { 
            ALL += balances[temp[i]][task_id];
        }
        uint share = ALL  / success.length;

		//给验证成功的TTP发钱     
        for (uint i = 0; i < success.length; i++) {
			TTP memory ttp = TTPS[success[i]];
            address payable recipient2 = payable(ttp.account);
            uint amount = balances[success[i]][task_id] + uint(ttp.EV_i)+ share;
            recipient2.transfer(amount);
        }
        address payable recipient3 = payable(tasks[task_id].date_owner);
        uint data_owner_fee = tasks[task_id].date_fee;
        recipient3.transfer(data_owner_fee);
        tasks[task_id].success_distribute1 = 1;   
        
    }

    //Allocation of Funds for Failed Task Executions
    function fail_distribute(uint task_id,uint[] memory success) public {
        //require(block.timestamp >= tasks[task_id].tasktime +  2 minutes, "Not enough time passed");
        require(success.length < tasks[task_id].n, "Record is already completed");
		tasks[task_id].fail_distribute1 == 1;
        uint[] memory temp = new uint[](tasks[task_id].sendersa.length-success.length);
		uint count = 0;
		for (uint i = 0; i < tasks[task_id].sendersa.length; i++) {
            bool found = false;
            for (uint j = 0; j < success.length; j++) {
                if (tasks[task_id].sendersa[i] == success[j]) {
                    found = true;
                    break;
                }
            }
            if (!found) {
				temp[count] = tasks[task_id].sendersa[i];
                count++;
            }
        }		
		//给验证失败的TTP发钱+质押后不提供数据的退钱
        for (uint i = 0; i < temp.length; i++) {
			TTP memory ttp = TTPS[temp[i]];
            address payable recipient1 = payable(ttp.account);
            uint refund = balances[temp[i]][task_id] * uint(ttp.RP_i) / 100;
            recipient1.transfer(refund);
            balances[temp[i]][task_id] -= refund;
        }
        uint ALL=0;
        for (uint i = 0; i < temp.length; i++) { 
            ALL += balances[temp[i]][task_id];
        }
        uint share = ALL  / (success.length+1);

		//给验证成功的TTP发钱     
        for (uint i = 0; i < success.length; i++) {
			TTP memory ttp = TTPS[success[i]];
            address payable recipient2 = payable(ttp.account);
            uint amount = balances[success[i]][task_id] + share;
            recipient2.transfer(amount);
        }
        address payable recipient3 = payable(tasks[task_id].date_user);
        uint data_user_fee = tasks[task_id].ALL_fee+share;
        recipient3.transfer(data_user_fee);
		tasks[task_id].fail_distribute1 == 1;
    }

    //Function to update CV_i
    function updateCY_i(uint task_id,uint[] memory success) public  {
        uint[] memory temp = new uint[](tasks[task_id].sendersa.length-success.length);
		uint count = 0;
		for (uint i = 0; i < tasks[task_id].sendersa.length; i++) {
            bool found = false;
            for (uint j = 0; j < success.length; j++) {
                if (tasks[task_id].sendersa[i] == success[j]) {
                    found = true;
                    break;
                }
            }
            if (!found) {
				temp[count] = tasks[task_id].sendersa[i];
                count++;
            }
        }		
        if (tasks[task_id].success_distribute1 == 1 ) {
            for (uint i = 0; i < success.length; i++) {
                TTPS[success[i]].CV_i += 5;
            }
            for (uint i = 0; i < temp.length; i++) {
                TTPS[temp[i]].CV_i -= 5;
            }
        }
        else if (tasks[task_id].fail_distribute1 == 1) {
            for (uint i = 0; i < success.length; i++) {
                TTPS[success[i]].CV_i += 10;
            }
            for (uint i = 0; i < temp.length; i++) {
                TTPS[temp[i]].CV_i -= 10;
            }
        }
    }

}
