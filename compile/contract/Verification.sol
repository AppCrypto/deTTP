pragma solidity ^0.8.0;
pragma experimental ABIEncoderV2;


// https://www.iacr.org/cryptodb/archive/2002/ASIACRYPT/50/50.pdf
contract Verification
{
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

	function HashToPoint(uint256 s)
        internal view returns (G1Point memory)
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
    function FindYforX(uint256 x)
        internal view returns (uint256, uint256)
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
	function P2() pure internal returns (G2Point memory) {
		return G2Point(
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
				return false;
			}
		}
		return true;
	}
	function DELQVerify(G1Point[] memory g,G1Point[] memory y1,G1Point[] memory h,G1Point[] memory y2,uint256 c,G1Point[] memory a1,G1Point[] memory a2,uint256[] memory z)public payable returns(bool)
	{
	  for(uint256 i=0;i<g.length;i++)
	  {
             G1Point memory gG=g1mul(g[i],z[i]);
             G1Point memory y1G=g1mul(y1[i],c);
             
             G1Point memory hG=g1mul(h[i],z[i]);
             G1Point memory y2G=g1mul(y2[i],c);
             
             if((a1[i].X!=g1add(gG,y1G).X)||(a1[i].Y!=g1add(gG,y1G).Y)||(a2[i].X!=g1add(hG,y2G).X)||(a2[i].Y!=g1add(hG,y2G).Y))
             {
               return false;
             }
	     return true;
	   }
	 }
	 
	 
	 G1Point[] CK;
	 
	 function UploadCK(G1Point[] memory _CK) public payable
	 {
	   for(uint256 i=0;i<_CK.length;i++)
	   {
	     CK.push(_CK[i]);
	   }
	 }
	 
	 function DownloadCK() public payable returns(G1Point[] memory)
	 {
	   return CK;
	 }
	 
	 G1Point[] EK_0;
	 G1Point[] EK_1;
	 
	 function UploadEK(G1Point[] memory _EK_0,G1Point[] memory _EK_1) public payable
	 {
	   for(uint256 i=0;i<_EK_0.length;i++)
	   {
	     EK_0.push(_EK_0[i]);
	     EK_1.push(_EK_1[i]);
	   }
	 }
	 
	 function DownloadEK_0() public payable returns(G1Point[] memory)
	 {
	   return EK_0;
	 }
	 
	 function DownloadEK_1() public payable returns(G1Point[] memory)
	 {
	   return EK_1;
	 }
	   
	 
	 G1Point[] Dis;
	 
	 
	 function UploadDispute(G1Point[] memory g,G1Point[] memory y1,G1Point[] memory h,G1Point[] memory y2,uint256 c,G1Point[] memory a1,G1Point[] memory a2,uint256[] memory z) public payable returns(bool)
	 {
	   if(!DELQVerify(g,y1,h,y2,c,a1,a2,z))
	   {
	     return false;
	   }
	   for(uint256 i=0;i<y2.length;i++)
	   {
	     Dis.push(y2[i]);
	   }
	   return true;
	 }	 	 
}
