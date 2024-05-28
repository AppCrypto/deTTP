// SPDX-License-Identifier: MIT
pragma solidity >=0.4.22 <0.9.0;

contract SimpleStorage {
    string  storedData;

    function set(string memory strx) public {
        storedData = strx;
    }

    function get() public view returns (string memory) {
        return storedData;
    }
}