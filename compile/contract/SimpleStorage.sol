// SPDX-License-Identifier: MIT
pragma solidity >=0.4.22 <0.9.0;

contract SimpleStorage {
    string  storedData;

    struct G1Point {
        uint256 X;
        uint256 Y;
    }

    G1Point G;


    function set(string memory strx) public {
        storedData = strx;
    }

    function get() public view returns (string memory) {
        return storedData;
    }

    function setG1Point(G1Point memory g) public{
        G=g;
    }

   function getG1Point() public view returns (G1Point memory) {
        return G;
    }
}
