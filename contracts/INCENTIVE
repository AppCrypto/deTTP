// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract INCENTIVE {
    
    struct TTP {
        int256 CV_i;      //credit value
        int256 EV_i;      //expected valu
        int256 RP_i;      //Refundable percentage*100
        int256 EDA_i;     //expected digital assets
    }
    int public MDA_i=50;  //minimum deposited assets
    int public a=6;
    int public b=3;
    int public success_distribute1=0;
    int public fail_distribute1=0;
    uint256 n=10;       //The number of ttp required to complete the task
    //Store information for each ttp
    mapping (address => TTP) public ttps;
    event Saved(address indexed sender, int256 CV_i, int256 EV_i, int256 RP_i, int256 EDA_i );
    
    uint public deployTime;

    uint256 public data_user;
    address public data_user_adress;
    // Define an event for recording transfer information
    event Transfer(address indexed from, address indexed to, uint256 value);

    // Define a constructor to initialize the timestamp of contract deployment
    constructor() payable {
        deployTime = block.timestamp;
    }
    
    address[] public senderss;
    //Store verified address
    address[] public senderst;
    //Store unverified addresses
    address[] public sendersf;
    // A mapping to store the ether balance of each user
    mapping(address => uint) public balances;
    address[] public sendersa;
    //data_user pay
    function user_pay() public payable {
       require(senderst.length == n, "ttp fee calculation not completed");   
       require(data_user == msg.value, "The amount you sent is wrong");
    }

    //Function to calculate EDA_i
    function save(int256 CV_i, int256 EV_i, int256 RP_i) public  {
       
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
    
        emit Saved(msg.sender, CV_i, EV_i, RP_i, EDA_i);
    }

    //Query TTP information
    function query() public view returns (int256, int256, int256, int256) {
        TTP memory ttp = ttps[msg.sender];
        return (ttp.CV_i, ttp.EV_i, ttp.RP_i, ttp.EDA_i);
    }
    

    // A function to deposit ether to the contract
    function deposit() public payable {
        TTP memory ttp = ttps[msg.sender];
        int256 A;
        A = ttp.EDA_i;
        uint256 B;
        B = balances[msg.sender];                
        require( B == 0, "You have already deposited");
        require(msg.value == uint256(A), "You must send  EDA_i wei");
        balances[msg.sender] += msg.value;
        sendersa.push(msg.sender);

    }
    //return funds function
    function withdraw(address payable recipient) public {
        // Check the balance of the recipient
        uint amount = balances[recipient];
        // Require that the recipient has a positive balance
        require(amount > 0, "You have no balance to withdraw");
        // Reset the balance of the recipient to zero
        balances[recipient] = 0;
        // Transfer the amount to the recipient
        recipient.transfer(amount);
    }  
    //ttp incoming verification message port
    function record(uint number) external {
        uint256 B;
        B = balances[msg.sender];                
        require(block.timestamp <= deployTime + 2 minutes, "Verification time exceeded");
        require( B != 0, "You must pledge funds");       
        if (number ==  10) {
            require(senderst.length < n, "Record is already completed");
            senderst.push(msg.sender);
            TTP memory ttp = ttps[msg.sender];
            data_user +=uint(ttp.EV_i);
        }
        else {
        sendersf.push(msg.sender);
        }
    }
    //Allocation of Funds for Successful Mission Execution
    function success_distribute() public {
        require(senderst.length == n, "The number of ttp has not reached the threshold");       
        for (uint i = 0; i < senderst.length; i++) {
            address payable recipient = payable(senderst[i]);
            TTP memory ttp = ttps[recipient];
            uint amount = balances[recipient] + uint(ttp.EV_i);
            recipient.transfer(amount);
        }
        for (uint i = 0; i < sendersf.length; i++) {
            address payable recipient = payable(sendersf[i]);
            TTP memory ttp = ttps[recipient];
            uint refund = balances[recipient] * uint(ttp.RP_i) / 100;
            recipient.transfer(refund);
        }
        uint remainingBalance = address(this).balance;
        uint share = remainingBalance  / senderst.length;
        for (uint i = 0; i < senderst.length; i++) {
            address payable recipient = payable(senderst[i]);
            recipient.transfer(share);
        }
        success_distribute1=1;
    }
    //Allocation of Funds for Failed Task Executions
    function fail_distribute() public {
        require(block.timestamp >= deployTime + 2 minutes, "Not enough time passed");
        require(senderst.length < n, "Record is already completed");
        for (uint i = 0; i < senderst.length; i++) {
            address payable recipient1 = payable(senderst[i]);
            withdraw(recipient1);
        }     
        for (uint i = 0; i < sendersa.length; i++) {
            address payable recipient2 = payable(sendersa[i]);
            TTP memory ttp = ttps[recipient2];
            uint refund = balances[recipient2] * uint(ttp.RP_i) / 100;
            recipient2.transfer(refund);
        }
        address payable recipient3 = payable(data_user_adress);
        recipient3.transfer(data_user);      
        uint remainingBalance = address(this).balance;
        uint share = remainingBalance / senderst.length;
        for (uint i = 0; i < senderst.length; i++) {
            address payable recipient4 = payable(senderst[i]);
            recipient4.transfer(share);
        }
        fail_distribute1=1;
    }

    //Function to update CV_i
    function updateCY_i() public  {
        if (success_distribute1 == 1 ) {
            for (uint i = 0; i < senderst.length; i++) {
                address recipient5 = senderst[i];
                ttps[recipient5].CV_i += 5;
            }
            for (uint i = 0; i < sendersf.length; i++) {
                address recipient5 = sendersf[i];
                ttps[recipient5].CV_i -= 5;
            }
        }
       
        else if (fail_distribute1 == 1) {
            for (uint i = 0; i < senderst.length; i++) {
                address recipient5 = senderst[i];
                ttps[recipient5].CV_i += 10;
            }
            for (uint i = 0; i < sendersf.length; i++) {
                address recipient5 = sendersa[i];
                ttps[recipient5].CV_i -= 5;
            }   
        }
    }
}
