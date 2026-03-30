// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IERC20} from "./openzeppelin-contracts/contracts/token/ERC20/IERC20.sol";
import {IERC20Permit} from "./openzeppelin-contracts/contracts/token/ERC20/extensions/IERC20Permit.sol";

/// @title PermitAndTransfer
/// @notice Pull tokens from `owner` to `to` in a single tx using EIP-2612 permit + transferFrom.
contract PermitAndTransfer {
    error TransferFromFailed();

    /// @notice Executes EIP-2612 permit and then transferFrom in a single call.
    /// @param token ERC20 token address (must implement EIP-2612).
    /// @param owner Token owner who signed the permit.
    /// @param to Recipient of the tokens.
    /// @param amount Amount to move.
    /// @param deadline Permit deadline (from the signed message).
    /// @param v Signature v
    /// @param r Signature r
    /// @param s Signature s
    function permitAndTransferFrom(
        address token,
        address owner,
        address to,
        uint256 amount,
        uint256 deadline,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) external {
        IERC20Permit(token).permit(owner, address(this), amount, deadline, v, r, s);

        bool ok = IERC20(token).transferFrom(owner, to, amount);
        if (!ok) revert TransferFromFailed();
    }
}