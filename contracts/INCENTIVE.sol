// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract INCENTIVE {
    
    struct TTP {
        int256 CV_i;      //credit value
        int256 EV_i;      //expected valu
        int256 RP_i;      //Refundable percentage*100
        int256 EDA_i;     //expected digital assets
    }
    struct task {
        uint  tasktime;
        address date_owner;     
        address date_user; 
        uint date_fee;    
        address[] TTPS;    
        uint256 n;       //The number of ttp required to complete the task
        address[] senderss;
        //Store verified address
        address[] senderst;
        //Store unverified addresses
        address[] sendersf;
        //Complete the ttp of the deposit funds.
        address[] sendersa;
        int success_distribute1;
        int fail_distribute1;
    }

    int public MDA_i=50;  //minimum deposited assets
    int public a=6;
    int public b=3;


    mapping (address => TTP) public ttps;
    mapping (address => task) public tasks;
    // A mapping to store the ether balance of each user
    mapping(address => mapping(address => uint)) public balances;

     function new_task(address date_owner, address date_user, uint date_fee, uint256 n) public  {
        tasks[date_user].tasktime = block.timestamp;
        tasks[date_user].date_owner = date_owner;
        tasks[date_user].date_user = date_user;
        tasks[date_user].date_fee = date_fee;
        tasks[date_user].n = n;
    }

    uint256 public ALL_fee=0;
    //date_user pay
    function date_user_fee(address date_user) public returns (uint256) {
       require(tasks[date_user].senderst.length == tasks[msg.sender].n, "ttp fee calculation not completed");   
       uint256 ALL_fees=0;
       for (uint i = 0; i < tasks[date_user].senderst.length; i++) {
            address ad = tasks[date_user].senderst[i];
            TTP memory ttp = ttps[ad];
            ALL_fees += uint(ttp.EV_i);
        }   
       ALL_fee = ALL_fees + tasks[date_user].date_fee;
       return (ALL_fees);
    }
    
      //date_user pay
    function date_user_pay(address date_user) public payable {
       require(tasks[date_user].senderst.length == tasks[msg.sender].n, "ttp fee calculation not completed");     
       //require(ALL_fee == msg.value, "The amount you sent is wrong");
    }


    //Function to calculate EDA_i
    function TTP_EDA_i(int256 CV_i, int256 EV_i, int256 RP_i) public  {
       
        int EDA_i;
        int A;

        A=a * EV_i * RP_i / 100 - b * CV_i;
        if (A >  MDA_i) {
            EDA_i = A;
        }
       
        else {
            EDA_i= MDA_i;
        }
        ttps[msg.sender] = TTP(CV_i, EV_i, RP_i, EDA_i );
    }

    //Query TTP information
    function query_TTP() public view returns (int256, int256, int256, int256) {
        TTP memory ttp = ttps[msg.sender];
        return (ttp.CV_i, ttp.EV_i, ttp.RP_i, ttp.EDA_i);
    }
    

    // A function to deposit ether to the contract
    function deposit(address date_user) public payable {
        TTP memory ttp = ttps[msg.sender];
        int256 A;
        A = ttp.EDA_i;
        uint256 B;
        B = balances[date_user][msg.sender];  
        require( B == 0, "You have already deposited");              
        require(msg.value == uint256(A), "You must send  EDA_i wei");
        balances[date_user][msg.sender] += msg.value;
        tasks[date_user].sendersa.push(msg.sender);
    }

 
    //ttp incoming verification message port
    function record(address date_user, uint number) external {
        uint256 B;
        B = balances[date_user][msg.sender];             
        require(block.timestamp <= tasks[date_user].tasktime + 1 minutes, "Verification time exceeded");
        require(tasks[date_user].senderst.length < tasks[date_user].n, "The number of ttp has not reached the threshold");       
        require( B != 0, "You must pledge funds");       
        if (number ==  10) {
            require(tasks[date_user].senderst.length < tasks[date_user].n, "Record is already completed");
            tasks[date_user].senderst.push(msg.sender);
            TTP memory ttp = ttps[msg.sender];
            tasks[date_user].date_fee +=uint(ttp.EV_i);
        }
        else {
        tasks[date_user].sendersf.push(msg.sender);
        }
    }


    //Allocation of Funds for Successful Mission Execution
    function success_distribute(address date_user) public  {
        require(tasks[date_user].senderst.length == tasks[date_user].n, "The number of ttp has not reached the threshold");       
        for (uint i = 0; i < tasks[date_user].senderst.length; i++) {
            address payable recipient3 = payable(tasks[date_user].senderst[i]);
            TTP memory ttp = ttps[recipient3];
            uint amount = balances[date_user][recipient3] + uint(ttp.EV_i);
            recipient3.transfer(amount);
        }
        for (uint i = 0; i < tasks[date_user].sendersf.length; i++) {
            address payable recipient4 = payable(tasks[date_user].sendersf[i]);
            TTP memory ttp = ttps[recipient4];
            uint refund = balances[date_user][recipient4] * uint(ttp.RP_i) / 100;
            recipient4.transfer(refund);
            balances[date_user][recipient4] -= refund;
        }
        uint ALL=0;
        for (uint i = 0; i < tasks[date_user].sendersf.length; i++) { 
            ALL += balances[date_user][tasks[date_user].sendersf[i]];
        }
        uint share = ALL  / tasks[date_user].senderst.length;
        for (uint i = 0; i < tasks[date_user].senderst.length; i++) {
            address payable recipient5 = payable(tasks[date_user].senderst[i]);
            recipient5.transfer(share);
        }
        address payable recipient6 = payable(tasks[date_user].date_owner);
        uint data_owner_fee = tasks[date_user].date_fee;
        recipient6.transfer(data_owner_fee);
        tasks[date_user].success_distribute1 = 1;   
        
    }

    //Allocation of Funds for Failed Task Executions
    function fail_distribute(address date_user) public {
        require(block.timestamp >= tasks[date_user].tasktime + 1 minutes, "Not enough time passed");
        require(tasks[date_user].senderst.length < tasks[date_user].n, "Record is already completed");
        for (uint i = 0; i < tasks[date_user].senderst.length; i++) {
            address payable recipient1 = payable(tasks[date_user].senderst[i]);
            TTP memory ttp = ttps[recipient1];
            uint amount = balances[date_user][recipient1];
            recipient1.transfer(amount);
            balances[date_user][recipient1] -= amount;
        }     
        for (uint i = 0; i < tasks[date_user].sendersa.length; i++) {
            address payable recipient2 = payable(tasks[date_user].sendersa[i]);
            TTP memory ttp = ttps[recipient2];
            uint refund = balances[date_user][recipient2] * uint(ttp.RP_i) / 100;
            recipient2.transfer(refund);
            balances[date_user][recipient2] -= refund;
        }
        uint ALL=0;
        for (uint i = 0; i < tasks[date_user].sendersa.length; i++) { 
            ALL += balances[date_user][tasks[date_user].sendersa[i]];
        }
        uint share = ALL / tasks[date_user].senderst.length;
        for (uint i = 0; i < tasks[date_user].senderst.length; i++) {
            address payable recipient4 = payable(tasks[date_user].senderst[i]);
            recipient4.transfer(share);
        }  

    }

    //Function to update CV_i
    function updateCY_i(address date_user) public  {
        if (tasks[date_user].success_distribute1 == 1 ) {
            for (uint i = 0; i < tasks[date_user].senderst.length; i++) {
                address recipient5 = tasks[date_user].senderst[i];
                ttps[recipient5].CV_i += 5;
            }
            for (uint i = 0; i < tasks[date_user].sendersf.length; i++) {
                address recipient5 = tasks[date_user].sendersf[i];
                ttps[recipient5].CV_i -= 5;
            }
        }
       
        else if (tasks[date_user].fail_distribute1 == 1) {
            for (uint i = 0; i < tasks[date_user].senderst.length; i++) {
                address recipient5 = tasks[date_user].senderst[i];
                ttps[recipient5].CV_i += 10;
            }
            for (uint i = 0; i < tasks[date_user].sendersf.length; i++) {
                address recipient5 = tasks[date_user].sendersa[i];
                ttps[recipient5].CV_i -= 5;
            }   
        }
    }
}
