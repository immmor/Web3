// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract MathOperations {
    // 加法运算
    function add(uint256 a, uint256 b) public pure returns (uint256) {
        return a + b;
    }
    
    // 减法运算
    function subtract(uint256 a, uint256 b) public pure returns (uint256) {
        require(b <= a, "Subtraction underflow");
        return a - b;
    }
}
    