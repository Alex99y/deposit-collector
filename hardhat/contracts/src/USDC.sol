// SPDX-License-Identifier: MIT 
pragma solidity ^0.8.24;

import {ERC3009} from "./base-contracts/ERC3009.sol";

contract USDC is ERC3009 {
    constructor() ERC3009("USDC", "USDC", "1.0.0") {
        _mint(msg.sender, 1000000000000000000000000);
    }
}