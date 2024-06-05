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


    // A mapping to store the ether balance of each task
    mapping(uint => mapping(uint => uint)) public balances;

		function new_task(address date_owner, address date_user, uint date_fee, uint256 n) public   returns (uint)
    {
         // Initializes a new Task object
        task memory newTask;
        newTask.tasktime = block.timestamp;
        newTask.date_owner = date_owner;
        newTask.date_user = date_user;
        newTask.date_fee = date_fee;
        newTask.n = n;
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

    //date_user fee
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

    function HashToPoint(uint256 s) internal view returns (G1Point memory)
    {
        uint256 beta = 0;
        uint256 y = 0;

        // XXX: Gen Order (n) or Field Order (p) ?
        uint256 x = s % CURVE_ORDER;

        while( true ) {
            (beta, y) = FindYforX(x);

            // y^2 == beta
            if( beta == mulmod(y, y, FIELD_ORDER) ) {
                return G1Point(x, y);
            }

            x = addmod(x, 1, FIELD_ORDER);
        }
    }


    /**
    * Given X, find Y
    *
    *   where y = sqrt(x^3 + b)
    *
    * Returns: (x^3 + b), y
	*/
 
    function FindYforX(uint256 x) internal view returns (uint256, uint256)
    {
        // beta = (x^3 + b) % p
        uint256 beta = addmod(mulmod(mulmod(x, x, FIELD_ORDER), x, FIELD_ORDER), CURVE_B, FIELD_ORDER);

        // y^2 = x^3 + b
        // this acts like: y = sqrt(beta)
        uint256 y = expMod(beta, CURVE_A, FIELD_ORDER);

        return (beta, y);
    }


    // a - b = c;
    function submod(uint a, uint b) internal pure returns (uint){
        uint a_nn;

        if(a>b) {
            a_nn = a;
        } else {
            a_nn = a+CURVE_ORDER;
        }

        return addmod(a_nn - b, 0, CURVE_ORDER);
    }


    function expMod(uint256 _base, uint256 _exponent, uint256 _modulus)
        internal view returns (uint256 retval)
    {
        bool success;
        uint256[1] memory output;
        uint[6] memory input;
        input[0] = 0x20;        // baseLen = new(big.Int).SetBytes(getData(input, 0, 32))
        input[1] = 0x20;        // expLen  = new(big.Int).SetBytes(getData(input, 32, 32))
        input[2] = 0x20;        // modLen  = new(big.Int).SetBytes(getData(input, 64, 32))
        input[3] = _base;
        input[4] = _exponent;
        input[5] = _modulus;
        assembly {
            success := staticcall(sub(gas(), 2000), 5, input, 0xc0, output, 0x20)
            // Use "invalid" to make gas estimation work
            //switch success case 0 { invalid }
        }
        require(success);
        return output[0];
    }


    /// return the generator of G2
    function P2() pure internal returns (G2Point memory) {return G2Point(
	[11559732032986387107991004021392285783925812861821192530917403151452391805634,
	10857046999023057135944570762232829481370756359578518086990519993285655852781],
	[4082367875863433681332203403145435568316851327593401208105741076214120093531,
	 8495653923123431417604973247489272438418190587263600148770280649306958101930]
	);
    }

    /// return the negation of p, i.e. p.add(p.negate()) should be zero.
    function g1neg(G1Point memory p) pure internal returns (G1Point memory) {
	// The prime q in the base field F_q for G1
	uint q = 21888242871839275222246405745257275088696311157297823662689037894645226208583;
	if (p.X == 0 && p.Y == 0)
		return G1Point(0, 0);
	return G1Point(p.X, q - (p.Y % q));
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
	require (success);
    }

    /// return the result of computing the pairing check
    /// e(p1[0], p2[0]) *  .... * e(p1[n], p2[n]) == 1
    /// For example pairing([P1(), P1().negate()], [P2(), P2()]) should
    /// return true.
    function pairing(G1Point[] memory p1, G2Point[] memory p2) view internal returns (bool) {
	require(p1.length == p2.length);
	uint elements = p1.length;
	uint inputSize = elements * 6;
	uint[] memory input = new uint[](inputSize);
	for (uint i = 0; i < elements; i++)
	{
	    input[i * 6 + 0] = p1[i].X;
	    input[i * 6 + 1] = p1[i].Y;
	    input[i * 6 + 2] = p2[i].X[0];
	    input[i * 6 + 3] = p2[i].X[1];
	    input[i * 6 + 4] = p2[i].Y[0];
	    input[i * 6 + 5] = p2[i].Y[1];
	}
	uint[1] memory out;
	bool success;
	assembly {
		    success := staticcall(sub(gas()	, 2000), 8, add(input, 0x20), mul(inputSize, 0x20), out, 0x20)
		    // Use "invalid" to make gas estimation work
		    //switch success case 0 { invalid }
	}
	require(success);
	return out[0] != 0;
    }

    /// Convenience method for a pairing check for two pairs.
    function pairingProd2(G1Point memory a1, G2Point memory a2, G1Point memory b1, G2Point memory b2) view internal returns (bool) {
	G1Point[] memory p1 = new G1Point[](2);
	G2Point[] memory p2 = new G2Point[](2);
	p1[0] = a1;
	p1[1] = b1;
	p2[0] = a2;
	p2[1] = b2;
	return pairing(p1, p2);
    }

    /// Convenience method for a pairing check for three pairs.
    function pairingProd3(
	G1Point memory a1, G2Point memory a2,
	G1Point memory b1, G2Point memory b2,
	G1Point memory c1, G2Point memory c2
    ) view internal returns (bool) {
	G1Point[] memory p1 = new G1Point[](3);
	G2Point[] memory p2 = new G2Point[](3);
	p1[0] = a1;
	p1[1] = b1;
	p1[2] = c1;
	p2[0] = a2;
	p2[1] = b2;
	p2[2] = c2;
	return pairing(p1, p2);
    }

    /// Convenience method for a pairing check for four pairs.
    function pairingProd4(
	G1Point memory a1, G2Point memory a2,
	G1Point memory b1, G2Point memory b2,
	G1Point memory c1, G2Point memory c2,
	G1Point memory d1, G2Point memory d2
	) view internal returns (bool) {
	G1Point[] memory p1 = new G1Point[](4);
	G2Point[] memory p2 = new G2Point[](4);
	p1[0] = a1;
	p1[1] = b1;
	p1[2] = c1;
	p1[3] = d1;
	p2[0] = a2;
	p2[1] = b2;
	p2[2] = c2;
	p2[3] = d2;
	return pairing(p1, p2);
    }

    // Costs ~85000 gas, 2x ecmul, + mulmod, addmod, hash etc. overheads
    function CreateProof( uint256 secret, uint256 message )
	public payable
	returns (uint256[2] memory out_pubkey, uint256 out_s, uint256 out_e)
    {
	G1Point memory xG = g1mul(P1(), secret % N());
	out_pubkey[0] = xG.X;
	out_pubkey[1] = xG.Y;
	uint256 k = uint256(keccak256(abi.encodePacked(message, secret))) % N();
	G1Point memory kG = g1mul(P1(), k);
	out_e = uint256(keccak256(abi.encodePacked(out_pubkey[0], out_pubkey[1], kG.X, kG.Y, message)));
	out_s = submod(k, mulmod(secret, out_e, N()));
    }

    // Costs ~85000 gas, 2x ecmul, 1x ecadd, + small overheads
    function CalcProof( uint256[2] memory pubkey, uint256 message, uint256 s, uint256 e )
	public payable
	returns (uint256)
    {
	G1Point memory sG = g1mul(P1(), s % N());
	G1Point memory xG = G1Point(pubkey[0], pubkey[1]);
	G1Point memory kG = g1add(sG, g1mul(xG, e));
	return uint256(keccak256(abi.encodePacked(pubkey[0], pubkey[1], kG.X, kG.Y, message)));
    }

    function modPow(uint256 base, uint256 exponent, uint256 modulus) internal returns (uint256) {
	uint256[6] memory input = [32,32,32,base,exponent,modulus];
	uint256[1] memory result;
	assembly {
	    if iszero(call(not(0), 0x05, 0, input, 0xc0, result, 0x20)) {
	    revert(0, 0)
	}
	    }
	    return result[0];
	}

	
    //The algotithm VerifyShare
    function VSSVerify(uint256[] memory arr, uint256 len1, uint256 len2)
	public payable
	returns (bool)
    {
	for(uint256 i=0;i<len1*2;i+=2) {
	    G1Point memory xG = g1mul(P1(), 0);
	    for(uint256 j=0;j<len2*2;j+=2) {
		uint256 seg =len1+2*len1+len2;
		G1Point memory comj = G1Point(arr[seg+j], arr[seg+j+1]);
		uint256 ipowj = modPow(arr[i/2],arr[len1+2*len1+j/2],CURVE_ORDER);
		xG = g1add(xG, g1mul(comj, ipowj)) ;
	    }
	    if(arr[len1+i] != xG.X || arr[len1+i+1]!= xG.Y){
		VerificationResult.push(false);
		return false;
	    }
	}
	VerificationResult.push(true);
	return true;
    }


    G1Point[]  g;
    G1Point[]  h;
    G1Point[]  y1;
    G1Point[]  y2;
    G1Point[]  a1;
    G1Point[]  a2;

    function Mul_DELQVerify(uint256[2][] memory G,uint256[2][] memory XG,uint256[2][] memory H,uint256[2][] memory XH,uint256[] memory c,uint256[2][] memory RG,uint256[2][] memory RH,uint256[] memory z) public payable returns(bool)
    {
	for (uint i = 0; i < G.length; i++) {
        G1Point memory Eg = G1Point(G[i][0], G[i][1]);
        g.push(Eg);
        G1Point memory Eh = G1Point(H[i][0], H[i][1]);
        h.push(Eh);
        G1Point memory Egx = G1Point(XG[i][0], XG[i][1]);
        y1.push(Egx);
        G1Point memory Ehx = G1Point(XH[i][0], XH[i][1]);
        y2.push(Ehx);
        G1Point memory Egr = G1Point(RG[i][0], RG[i][1]);
        a1.push(Egr);
        G1Point memory Ehr = G1Point(RH[i][0], RH[i][1]);
        a2.push(Ehr);
        }

	for(uint256 i=0;i<g.length;i++){
	    G1Point memory gG=g1mul(g[i],z[i]);
            G1Point memory y1G=g1mul(y1[i],c[i]);
            G1Point memory hG=g1mul(h[i],z[i]);
            G1Point memory y2G=g1mul(y2[i],c[i]); 
            if((a1[i].X!=g1add(gG,y1G).X)||(a1[i].Y!=g1add(gG,y1G).Y)||(a2[i].X!=g1add(hG,y2G).X)||(a2[i].Y!=g1add(hG,y2G).Y)){
            	VerificationResult.push(true);
		return false;
            }
	}
	VerificationResult.push(true);
	return true;
    }

    function DELQVerify(uint256[2] memory G,uint256[2] memory XG,uint256[2] memory H,uint256[2] memory XH,uint256 c,uint256[2] memory RG,uint256[2] memory RH,uint256 z) public payable 
    {
        G1Point memory _g = G1Point(G[0], G[1]);
        G1Point memory _h = G1Point(H[0], H[1]);
        G1Point memory _y1 = G1Point(XG[0], XG[1]);
        G1Point memory _y2 = G1Point(XH[0], XH[1]);
		G1Point memory _a1 = G1Point(RG[0], RG[1]);
        G1Point memory _a2 = G1Point(RH[0], RH[1]);
 
		G1Point memory gG=g1mul(_g,z);
        G1Point memory y1G=g1mul(_y1,c);
             
       	G1Point memory hG=g1mul(_h,z);
        G1Point memory y2G=g1mul(_y2,c);
             
        if((_a1.X!=g1add(gG,y1G).X)||(_a1.Y!=g1add(gG,y1G).Y)||(_a2.X!=g1add(hG,y2G).X)||(_a2.Y!=g1add(hG,y2G).Y))
        {
			VerificationResult.push(false);
        }
		VerificationResult.push(true);
   }

   struct Ciphertext{
	G1Point C0;
	G1Point C1;
    }

    struct DLEQProof{
	G1Point g;
	G1Point h;
	G1Point y1;
	G1Point y2;
	G1Point a1;
	G1Point a2;
	uint256 c;
	uint256 z;
    }

    struct EKey{
	G1Point[] EK0;
	G1Point[] EK1;
    }
	
    G1Point[] owner_pk;
    G1Point[] user_pk;
    G1Point[] TPPs_pk;
    Ciphertext[] Ciphertexts;
    G1Point[] Gs;
    G1Point[] Commitments;
    G1Point[] CKeys;
    DLEQProof[] DLEQProofs;
    EKey[] EKeys;

    function upload_owner_pk(uint256[2] memory pk)public {
	G1Point memory E = G1Point(pk[0],pk[1]);
	owner_pk.push(E);
    }

    function upload_user_pk(uint256[2] memory pk)public {
	G1Point memory E = G1Point(pk[0],pk[1]);
	user_pk.push(E);
    }

    function upload_TTPs_pk(uint256[2] memory pk)public {
	G1Point memory E = G1Point(pk[0],pk[1]);
	owner_pk.push(E);
    }

    // 新增函数：一次上传多个TTP_pk
    function upload_multiple_TTP_pk(uint256[2][] memory pkArray) public {
        for (uint i = 0; i < pkArray.length; i++) {
            G1Point memory E = G1Point(pkArray[i][0], pkArray[i][1]);
            TPPs_pk.push(E);
        }
    }

    function upload_ciphertext(uint256[2] memory c0, uint256[2] memory c1)public {
	G1Point memory E0 = G1Point(c0[0],c0[1]);
	G1Point memory E1 = G1Point(c1[0],c1[1]);
	Ciphertext memory ciphertext;
	ciphertext.C0=E0;
	ciphertext.C1=E1;
	Ciphertexts.push(ciphertext);
    }

    function GsAndCommitment(uint256[2][] memory gs, uint256[2][] memory commitments)public {
	for (uint i = 0; i < gs.length; i++) {
            G1Point memory g = G1Point(gs[i][0], gs[i][1]);
            Gs.push(g);
        }
	G1Point memory c;
	for (uint i = 0; i < commitments.length; i++) {
            c = G1Point(commitments[i][0], commitments[i][1]);
            Commitments.push(c);
        }
    }

    function upload_CKey(uint256[2][] memory ckeys) public {
        for (uint i = 0; i < CKeys.length; i++) {
            G1Point memory ckey = G1Point(ckeys[i][0],ckeys[i][1]);
            CKeys.push(ckey);
        }
    }

    function upload_DLEQProof(uint256[2][] memory G,uint256[2][] memory XG,uint256[2][] memory H,uint256[2][] memory XH,uint256[] memory C,uint256[2][] memory RG,uint256[2][] memory RH,uint256[] memory Z) public {
	for (uint i = 0; i < G.length; i++) {
            G1Point memory Eg = G1Point(G[i][0], G[i][1]);
            DLEQProofs[i].g=Eg;
            G1Point memory Eh = G1Point(H[i][0], H[i][1]);
            DLEQProofs[i].h=Eh;
            G1Point memory Egx = G1Point(XG[i][0], XG[i][1]);
            DLEQProofs[i].y1=Egx;
            G1Point memory Ehx = G1Point(XH[i][0], XH[i][1]);
            DLEQProofs[i].y2=Ehx;
            G1Point memory Egr = G1Point(RG[i][0], RG[i][1]);
            DLEQProofs[i].a1=Egr;
            G1Point memory Ehr = G1Point(RH[i][0], RH[i][1]);
            DLEQProofs[i].a2=Ehr;
	    DLEQProofs[i].c=C[i];
	    DLEQProofs[i].z=Z[i];
	}
    }

    G1Point[] ekeys0;
    G1Point[] ekeys1;
    EKey ekeys;
    function upload_EKey(uint256[2][] memory EKeys0,uint256[2][] memory EKeys1) public{
	for (uint i = 0; i < EKeys0.length; i++){
	    G1Point memory E0 = G1Point(EKeys0[i][0], EKeys0[i][1]);
	    ekeys0[i]=E0;
	    G1Point memory E1 = G1Point(EKeys1[i][0], EKeys1[i][1]);
	    ekeys1[i]=E1;
	}
	ekeys.EK0=ekeys0;
	ekeys.EK1=ekeys1;
	EKeys.push(ekeys);
    }

    bool[] VerificationResult;
    function get() public view returns (bool[] memory) {
        return VerificationResult;
    }
}
