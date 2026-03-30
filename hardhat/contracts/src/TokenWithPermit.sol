// SPDX-License-Identifier: MIT 
pragma solidity ^0.8.24;

import {ERC20} from "../openzeppelin-contracts/contracts/token/ERC20/ERC20.sol";
import {ERC20Permit} from "../openzeppelin-contracts/contracts/token/ERC20/extensions/ERC20Permit.sol";

contract KKK is ERC20Permit {
    constructor() ERC20("KKK", "KKK") ERC20Permit("KKK") {
        _mint(msg.sender, 1000000000000000000000000);
    }
}