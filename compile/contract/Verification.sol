// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

// https://www.iacr.org/cryptodb/archive/2002/ASIACRYPT/50/50.pdf
contract Verification
{
    //Storing TTP information
    struct TTP {
        int256 CV_i;      //credit value
        int256 EV_i;      //expected valu
        int256 RP_i;      //Refundable percentage*100
        int256 EDA_i;     //expected digital assets
        address account;
    }

    struct task {
        uint tasktime;
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


    int public MDA_i = 50;  //minimum deposited assets
    int public a = 6;
    int public b = 3;

    // A mapping to store the ether balance of each user
    mapping(uint => mapping(uint => uint)) public balances;
    mapping(uint => uint[]) public task_success;
    mapping(uint => uint[]) public task_failed;

    function new_task(address date_owner, address date_user, uint date_fee, uint256 n) public returns (uint)
    {

        task memory newTask;
        newTask.tasktime = block.timestamp;
        newTask.date_owner = date_owner;
        newTask.date_user = date_user;
        newTask.date_fee = date_fee;
        newTask.n = n;


        newTask.TTPS = new address[](0);
        newTask.sendersdata = new uint[](0);
        newTask.sendersa = new uint[](0);
        newTask.success_distribute1 = 0;
        newTask.fail_distribute1 = 0;
        newTask.ALL_fee = 0;
        tasks.push(newTask);
        return tasks.length - 1;
    }

    //Function to calculate EDA_i
    function TTP_register(int256 CV_i, int256 EV_i, int256 RP_i, address account) public returns (uint, int){

        int EDA_i;
        int A;

        A = a * EV_i * RP_i / 100 - b * CV_i;
        if (A > MDA_i) {
            EDA_i = A;
        }

        else {
            EDA_i = MDA_i;
        }
        uint id = TTPS.length;
        TTPS.push(TTP(CV_i, EV_i, RP_i, EDA_i, account));
        return (id, EDA_i);
    }
    //Query TTP information
    function query_TTP(uint id) public view returns (int256, int256, int256, int256, address) {
        return (TTPS[id].CV_i, TTPS[id].EV_i, TTPS[id].RP_i, TTPS[id].EDA_i, TTPS[id].account);
    }

    // A function to deposit ether to the contract
    function deposit(uint TTP_id, uint task_id) public payable {
        TTP memory ttp = TTPS[TTP_id];
        int256 A = ttp.EDA_i;
        uint256 B;
        B = balances[TTP_id][task_id];
        require(B == 0, "You have already deposited");
        require(msg.value == uint256(A), "You must send  EDA_i wei");
        balances[TTP_id][task_id] = msg.value;
        tasks[task_id].sendersa.push(TTP_id);
    }

    //date_user pay
    function date_user_fee(uint task_id) public returns (uint256) {
        uint256 ALL_fees = 0;
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

    function Submit_verification_results(uint task_id, uint[] memory success, uint[] memory failed) public {
        task_success[task_id] = success;
        task_failed[task_id] = failed;
    }

    //Allocation of Funds for Successful Mission Execution
    function success_distribute(uint task_id) public {
        require(block.timestamp <= tasks[task_id].tasktime + 2 minutes, "Not enough time passed");
        uint[] memory success = task_success[task_id];
        uint[] memory failed = task_failed[task_id];
        require(success.length >= tasks[task_id].n, "The number of ttp has not reached the threshold");

        for (uint i = 0; i < failed.length; i++) {
            TTP memory ttp = TTPS[failed[i]];
            address payable recipient1 = payable(ttp.account);
            uint refund = balances[failed[i]][task_id] * uint(ttp.RP_i) / 100;
            recipient1.transfer(refund);
            balances[failed[i]][task_id] -= refund;
        }
        uint ALL = 0;
        for (uint i = 0; i < failed.length; i++) {
            ALL += balances[failed[i]][task_id];
        }
        uint share = ALL / success.length;

        for (uint i = 0; i < success.length; i++) {
            TTP memory ttp = TTPS[success[i]];
            address payable recipient2 = payable(ttp.account);
            uint amount = balances[success[i]][task_id] + uint(ttp.EV_i) + share;
            recipient2.transfer(amount);
        }
        address payable recipient3 = payable(tasks[task_id].date_owner);
        uint data_owner_fee = tasks[task_id].date_fee;
        recipient3.transfer(data_owner_fee);
        tasks[task_id].success_distribute1 = 1;

    }

    //Allocation of Funds for Failed Task Executions
    function fail_distribute(uint task_id) public {
        //require(block.timestamp >= tasks[task_id].tasktime +  2 minutes, "Not enough time passed");
        uint[] memory success = task_success[task_id];
        uint[] memory failed = task_failed[task_id];
        require(success.length < tasks[task_id].n, "Record is already completed");
        for (uint i = 0; i < failed.length; i++) {
            TTP memory ttp = TTPS[failed[i]];
            address payable recipient1 = payable(ttp.account);
            uint refund = balances[failed[i]][task_id] * uint(ttp.RP_i) / 100;
            recipient1.transfer(refund);
            balances[failed[i]][task_id] -= refund;
        }
        uint ALL = 0;
        for (uint i = 0; i < failed.length; i++) {
            ALL += balances[failed[i]][task_id];
        }
        uint share = ALL / (success.length + 1);

        //给验证成功的TTP发钱
        for (uint i = 0; i < success.length; i++) {
            TTP memory ttp = TTPS[success[i]];
            address payable recipient2 = payable(ttp.account);
            uint amount = balances[success[i]][task_id] + share;
            recipient2.transfer(amount);
        }
        address payable recipient3 = payable(tasks[task_id].date_user);
        uint data_user_fee = tasks[task_id].ALL_fee + share;
        recipient3.transfer(data_user_fee);
        tasks[task_id].fail_distribute1 == 1;
    }

    //Function to update CV_i
    function updateCY_i(uint task_id) public {
        uint[] memory success = task_success[task_id];
        uint[] memory failed = task_failed[task_id];
        if (tasks[task_id].success_distribute1 == 1) {
            for (uint i = 0; i < success.length; i++) {
                TTPS[success[i]].CV_i += 5;
            }
            for (uint i = 0; i < failed.length; i++) {
                TTPS[failed[i]].CV_i -= 2;
            }
        }
        else if (tasks[task_id].fail_distribute1 == 1) {
            for (uint i = 0; i < success.length; i++) {
                TTPS[success[i]].CV_i += 5;
            }
            for (uint i = 0; i < failed.length; i++) {
                TTPS[failed[i]].CV_i -= 5;
            }
        }
    }

    mapping(string => address) public id2Addrs;

    function register(string memory id)
    public
    payable
    returns (bool)
    {
        id2Addrs[id] = msg.sender;
        return true;
    }

    // p = p(u) = 36u^4 + 36u^3 + 24u^2 + 6u + 1
    uint256 constant FIELD_ORDER = 0x30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47;

    // Number of elements in the field (often called `q`)
    // n = n(u) = 36u^4 + 36u^3 + 18u^2 + 6u + 1
    //uint256 constant CURVE_ORDER = 0x30644e72e131a029b85045b68181585d2833e84879b9709143e1f593f0000001;
    uint256 constant CURVE_ORDER = 0x30644e72e131a029b85045b68181585d2833e84879b9709143e1f593f0000001;
    uint256 constant CURVE_B = 3;

    // a = (p+1) / 4
    uint256 constant CURVE_A = 0xc19139cb84c680a6e14116da060561765e05aa45a1c72a34f082305b61f3f52;

    struct G1Point {
        uint X;
        uint Y;
    }

    // Encoding of field elements is: X[0] * z + X[1]
    struct G2Point {
        uint[2] X;
        uint[2] Y;
    }

    // (P+1) / 4
    function A() pure internal returns (uint256) {
        return CURVE_A;
    }

    function P() pure internal returns (uint256) {
        return FIELD_ORDER;
    }

    function N() pure internal returns (uint256) {
        return CURVE_ORDER;
    }

    /// return the generator of G1
    function P1() pure internal returns (G1Point memory) {
        return G1Point(1, 2);
    }

    /// return the sum of two points of G1
    function g1add(G1Point memory p1, G1Point memory p2) view internal returns (G1Point memory r) {
        uint[4] memory input;
        input[0] = p1.X;
        input[1] = p1.Y;
        input[2] = p2.X;
        input[3] = p2.Y;
        bool success;
        assembly {
            success := staticcall(sub(gas(), 2000), 6, input, 0xc0, r, 0x60)
        // Use "invalid" to make gas estimation work
        //switch success case 0 { invalid }
        }
        require(success);
    }

    /// return the product of a point on G1 and a scalar, i.e.
    /// p == p.mul(1) and p.add(p) == p.mul(2) for all points p.
    function g1mul(G1Point memory p, uint s) view internal returns (G1Point memory r) {
        uint[3] memory input;
        input[0] = p.X;
        input[1] = p.Y;
        input[2] = s;
        bool success;
        assembly {
            success := staticcall(sub(gas(), 2000), 7, input, 0x80, r, 0x60)
        // Use "invalid" to make gas estimation work
        //switch success case 0 { invalid }
        }
        require(success);
    }

    function VSSVerify(G1Point[] memory commitments) public payable returns (bool)
    {
        for(uint i=0;i<Gs.length;i++){
            G1Point memory right;
            uint256 xPower=1;
            uint256 x=i+1;
            G1Point memory temp;
            for(uint j=0;j<commitments.length;j++){
                temp = g1mul(commitments[j], xPower);
                right=g1add(right, temp);
                xPower=mulmod(xPower, x, CURVE_ORDER);
            }
            if(Gs[i].X != right.X||Gs[i].Y != right.Y){
                VerificationResult.push(false);
                return false;
            }
        }
        VerificationResult.push(true);
        return true;
    }

    function DLEQVerifyCKeys() public payable returns (bool)
    {
        for (uint256 i = 0; i < DLEQProofCKeys.length; i++)
        {
            G1Point memory gG = g1mul(g, DLEQProofCKeys[i].z);
            G1Point memory y1G = g1mul(Gs[i], DLEQProofCKeys[i].c);

            G1Point memory hG = g1mul(g1add(ciphertext.C0, TTPsPk[i]), DLEQProofCKeys[i].z);
            G1Point memory y2G = g1mul(CKeys[i], DLEQProofCKeys[i].c);

            G1Point memory pt1 =  g1add(gG, y1G);
            G1Point memory pt2 =  g1add(hG, y2G);
            if ((DLEQProofCKeys[i].a1.X != pt1.X) || (DLEQProofCKeys[i].a1.Y != pt1.Y) || (DLEQProofCKeys[i].a2.X != pt2.X) || (DLEQProofCKeys[i].a2.Y != pt2.Y))
            {
                VerificationResult.push(false);
                return false;
            }
        }
        VerificationResult.push(true);
        return true;
    }

    function DLEQVerifyDis(uint index) public payable returns (bool)
    {
        G1Point memory gG = g1mul(g, DLEQProofDis.z);
        G1Point memory y1G = g1mul(UserPk, DLEQProofDis.c);
        G1Point memory hG = g1mul(EKeys[index].EK0, DLEQProofDis.z);
        G1Point memory y2G = g1mul(Dispute, DLEQProofDis.c);

        G1Point memory pt1 =  g1add(gG, y1G);
        G1Point memory pt2 =  g1add(hG, y2G);
        if ((DLEQProofDis.a1.X != pt1.X) || (DLEQProofDis.a1.Y != pt1.Y) || (DLEQProofDis.a2.X != pt2.X) || (DLEQProofDis.a2.Y != pt2.Y))
        {
            VerificationResult.push(false);
            return false;
        }
        VerificationResult.push(true);
        return true;
    }

    // Encoding of field elements is: X[0] * z + X[1]


    struct Ciphertext {
        G1Point C0;
        G1Point C1;
    }

    struct DLEQProof {
        G1Point a1;
        G1Point a2;
        uint256 c;
        uint256 z;
    }

    struct EKey {
        G1Point EK0;
        G1Point EK1;
    }

    G1Point g;
    G1Point OwnerPk;
    G1Point UserPk;
    G1Point[] TTPsPk;
    Ciphertext ciphertext;
    G1Point[] Gs;
    G1Point[] CKeys;
    DLEQProof[] DLEQProofCKeys;
    DLEQProof[] DLEQProofKeys;
    G1Point Dispute;
    DLEQProof DLEQProofDis;
    EKey[] EKeys;
    bool[] VerificationResult;

    function UploadGenerator(G1Point memory generator) public {
        g = generator;
    }

    function UploadOwnerPk(G1Point memory pk) public {
        OwnerPk = pk;
    }

    function UploadUserPk(G1Point memory pk) public {
        UserPk = pk;
    }

    function UploadTTPsPk(G1Point[] memory pkArray) public {
        //G1Point[] memory TPPs_PKs = new G1Point[](pkArray.length);
        for (uint i = 0; i < pkArray.length; i++) {
            TTPsPk.push(pkArray[i]);
        }
    }

    function UploadCiphertext(G1Point memory c0, G1Point memory c1) public {
        ciphertext.C0 = c0;
        ciphertext.C1 = c1;
    }


    function UploadGs(G1Point[] memory gs) public {
        for (uint i = 0; i < gs.length; i++) {
            Gs.push(gs[i]);
        }
    }

    function UploadCKeys(G1Point[] memory ckeys) public {
        //G1Point[] memory CKeys = new G1Point[](ckeys.length);
        for (uint i = 0; i < ckeys.length; i++) {
            CKeys.push(ckeys[i]);
        }
    }

    function UploadDLEQProofCKeys(uint256[] memory c, G1Point[] memory a1, G1Point[] memory a2, uint256[] memory z) public {
        DLEQProof memory DLEQProofCKey;
        for (uint i = 0; i < c.length; i++) {
            DLEQProofCKey.c=c[i];
            DLEQProofCKey.a1=a1[i];
            DLEQProofCKey.a2=a2[i];
            DLEQProofCKey.z=z[i];
            DLEQProofCKeys.push(DLEQProofCKey);
        }
    }

    function UploadDLEQProofKeys(uint256[] memory _c, G1Point[] memory _a1, G1Point[] memory _a2, uint256[] memory _z) public {
        DLEQProof memory DLEQProofKey;
        for (uint i = 0; i < _c.length; i++) {
            DLEQProofKey.a1 = _a1[i];
            DLEQProofKey.a2 = _a2[i];
            DLEQProofKey.c = _c[i];
            DLEQProofKey.z = _z[i];
            DLEQProofKeys.push(DLEQProofKey);
        }
    }
 
    function UploadDispute(G1Point memory Dis) public {
        Dispute=Dis;
    }

    function UploadDisputeProof(uint256 c, G1Point memory a1, G1Point memory a2, uint256 z) public {
        DLEQProofDis.a1 = a1;
        DLEQProofDis.a2 = a2;
        DLEQProofDis.c = c;
        DLEQProofDis.z = z;
    }

    function UploadEKeys(G1Point[] memory EKeys0, G1Point[] memory EKeys1) public {
       EKey memory eKey;
        for (uint i = 0; i < EKeys0.length; i++) {
            eKey.EK0 = EKeys0[i];
            eKey.EK1 = EKeys1[i];
            EKeys.push(eKey);
        }
    }

    function GetVrfResult() public view returns (bool[] memory) {
        return VerificationResult;
    }

}
