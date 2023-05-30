// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract INC {
    // 声明一个结构体，用来存储两个数字
    struct TTP {
        int256 CV_i;
        int256 EV_i;
        int256 EDA_i;
    }
    int public MDA_i=50;
    int public a=6;
    int public b=3;
    // 声明一个映射，用来将地址映射到结构体
    mapping (address => TTP) public ttps;

    // 声明一个事件，用来保存TTP信息
    event Saved(address indexed sender, int256 CV_i, int256 EV_i, int256 EDA_i);

    // 根据TTP的信誉值与自己预期奖励计算质押资金
    function save(int256 CV_i, int256 EV_i) public  {
       
        int EDA_i;
        int A;

        A=a * EV_i - b * CV_i;
        if (A >  MDA_i) {
            EDA_i = A;
        }
       
        else {
            EDA_i= MDA_i;
        }

        ttps[msg.sender] = TTP(CV_i, EV_i, EDA_i );
    
        emit Saved(msg.sender, CV_i, EV_i, EDA_i);
    }

    struct User {
        address userAddress;
        uint256 amount;
    }
        
    // 查询TTP的信息
    function query() public view returns (int256, int256, int256) {
        TTP memory ttp = ttps[msg.sender];
        return (ttp.CV_i,ttp.EV_i,ttp.EDA_i);
    }

    
    mapping(address => User) public users;

    // 存放所有用户
    User[] public userList;

    // 存放所有资金数量
    uint256 public totalFunds;

    // 储存事件S
    bool public eventSCompleted;

    // 检查事件S是否完成
    modifier onlyEventSCompleted() {
        require(eventSCompleted, "Event S is not completed");
        _;
    }

    // 检查发送者是否为外部合约
    modifier onlyExternalContract() {
        require(msg.sender == externalContract, "Only external contract can call this function");
        _;
    }

    // 储存外部合约地址
    address public externalContract;

    // 设置外部合约地址的构造函数
    constructor(address _externalContract) {
        externalContract = _externalContract;
    }

    // 通知用户事件 S 已完成的事件
    event EventSCompleted();

    // 一个函数来存放资金
    function deposit() public payable {
        // Check that the sender sends a positive amount
        TTP memory ttp = ttps[msg.sender];
        require(int256(msg.value) > ttp.EDA_i, "Amount must be positive");
        // Check that event S is not completed
        require(!eventSCompleted, "Event S is completed");
        // Update the user's amount
        users[msg.sender].amount += msg.value;
        // Update the user's address
        users[msg.sender].userAddress = msg.sender;
        // Update the total amount of funds
        totalFunds += msg.value;
        // Add the user to the array if not already in
        if (users[msg.sender].amount == msg.value) {
            userList.push(users[msg.sender]);
        }
    }

    // 通知合约外部合约完成事件S的函数
    function notifyEventSCompleted() public onlyExternalContract {
 
        require(!eventSCompleted, "Event S is already completed");

        eventSCompleted = true;

        emit EventSCompleted();
    }
    //返还TTP资金，没收作恶节点资金，奖励诚实节点
    function refundAll() public onlyEventSCompleted {

        require(address(this).balance >= totalFunds, "Insufficient balance in contract");
        uint256 refundFunds = 0;

        for (uint256 i = 0; i < userList.length; i++) {
            if (inArrayA(userList[i].userAddress)) {
                uint256 amount = userList[i].amount;
                userList[i].amount = 0;
                payable(userList[i].userAddress).transfer(amount);
                refundFunds += amount;
            }
        }
        uint256 remainingFunds = totalFunds - refundFunds;
        uint256 numUsers = arrayA.length;
        uint256 averageFunds = remainingFunds / numUsers;
        for (uint256 i = 0; i < arrayA.length; i++) {
            payable(arrayA[i]).transfer(averageFunds);
        }
    }

    
    address[] public arrayA;

    // 检查一个地址是否在A内
    function inArrayA(address _address) public view returns (bool) {
        for (uint256 i = 0; i < arrayA.length; i++) {
            if (arrayA[i] == _address) {
                // Return true
                return true;
            }
        }
        // Return false
        return false;
    }

}
