// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {ERC20} from "../openzeppelin-contracts/contracts/token/ERC20/ERC20.sol";
import {EIP712} from "../openzeppelin-contracts/contracts/utils/cryptography/EIP712.sol";
import {ECDSA} from "../openzeppelin-contracts/contracts/utils/cryptography/ECDSA.sol";

/**
 * @dev Interface for https://eips.ethereum.org/EIPS/eip-3009[ERC-3009] transfer with authorization.
 */
interface IERC3009 {
    enum AuthorizationState {
        Unused,
        Used,
        Canceled
    }

    event AuthorizationUsed(address indexed authorizer, bytes32 indexed nonce);
    event AuthorizationCanceled(address indexed authorizer, bytes32 indexed nonce);

    function transferWithAuthorization(
        address from,
        address to,
        uint256 value,
        uint256 validAfter,
        uint256 validBefore,
        bytes32 nonce,
        bytes calldata signature
    ) external;

    function receiveWithAuthorization(
        address from,
        address to,
        uint256 value,
        uint256 validAfter,
        uint256 validBefore,
        bytes32 nonce,
        bytes calldata signature
    ) external;

    function cancelAuthorization(address authorizer, bytes32 nonce, bytes calldata signature) external;

    function authorizationState(address authorizer, bytes32 nonce) external view returns (AuthorizationState);
}

/**
 * @dev Abstract ERC-20 extension implementing EIP-712 signed authorizations for gasless transfers.
 *
 * Concrete tokens should inherit this contract and provide {ERC20} name/symbol and an EIP-712 `version`
 * string (e.g. `"1"`) suitable for your deployment domain.
 */
abstract contract ERC3009 is ERC20, EIP712, IERC3009 {
    bytes32 private constant _TRANSFER_WITH_AUTHORIZATION_TYPEHASH =
        keccak256(
            "TransferWithAuthorization(address from,address to,uint256 value,uint256 validAfter,uint256 validBefore,bytes32 nonce)"
        );

    bytes32 private constant _RECEIVE_WITH_AUTHORIZATION_TYPEHASH =
        keccak256(
            "ReceiveWithAuthorization(address from,address to,uint256 value,uint256 validAfter,uint256 validBefore,bytes32 nonce)"
        );

    bytes32 private constant _CANCEL_AUTHORIZATION_TYPEHASH =
        keccak256("CancelAuthorization(address authorizer,bytes32 nonce)");

    mapping(address authorizer => mapping(bytes32 nonce => AuthorizationState)) private _authorizationStates;

    error ERC3009InvalidAuthorization();
    error ERC3009AuthorizationNotYetValid();
    error ERC3009AuthorizationExpired();
    error ERC3009InvalidSigner();
    error ERC3009CallerNotPayee();
    error ERC3009ZeroAddress();

    constructor(string memory name_, string memory symbol_, string memory version_) ERC20(name_, symbol_) EIP712(name_, version_) {}

    /// @inheritdoc IERC3009
    function authorizationState(address authorizer, bytes32 nonce) public view override returns (AuthorizationState) {
        return _authorizationStates[authorizer][nonce];
    }

    /// @inheritdoc IERC3009
    function transferWithAuthorization(
        address from,
        address to,
        uint256 value,
        uint256 validAfter,
        uint256 validBefore,
        bytes32 nonce,
        bytes calldata signature
    ) external override {
        _requireValidAuthorization(from, validAfter, validBefore, nonce);
        _requireTransferSignature(
            from,
            to,
            value,
            validAfter,
            validBefore,
            nonce,
            _TRANSFER_WITH_AUTHORIZATION_TYPEHASH,
            signature
        );
        _markAuthorizationUsed(from, nonce);
        _transfer(from, to, value);
    }

    /// @inheritdoc IERC3009
    function receiveWithAuthorization(
        address from,
        address to,
        uint256 value,
        uint256 validAfter,
        uint256 validBefore,
        bytes32 nonce,
        bytes calldata signature
    ) external override {
        if (msg.sender != to) {
            revert ERC3009CallerNotPayee();
        }
        _requireValidAuthorization(from, validAfter, validBefore, nonce);
        _requireTransferSignature(
            from,
            to,
            value,
            validAfter,
            validBefore,
            nonce,
            _RECEIVE_WITH_AUTHORIZATION_TYPEHASH,
            signature
        );
        _markAuthorizationUsed(from, nonce);
        _transfer(from, to, value);
    }

    /// @inheritdoc IERC3009
    function cancelAuthorization(address authorizer, bytes32 nonce, bytes calldata signature) external override {
        if (authorizer == address(0)) {
            revert ERC3009ZeroAddress();
        }
        AuthorizationState state = _authorizationStates[authorizer][nonce];
        if (state != AuthorizationState.Unused) {
            revert ERC3009InvalidAuthorization();
        }
        bytes32 structHash = keccak256(abi.encode(_CANCEL_AUTHORIZATION_TYPEHASH, authorizer, nonce));
        address signer = ECDSA.recoverCalldata(_hashTypedDataV4(structHash), signature);
        if (signer != authorizer) {
            revert ERC3009InvalidSigner();
        }
        _authorizationStates[authorizer][nonce] = AuthorizationState.Canceled;
        emit AuthorizationCanceled(authorizer, nonce);
    }

    function _requireValidAuthorization(address authorizer, uint256 validAfter, uint256 validBefore, bytes32 nonce) private view {
        if (authorizer == address(0)) {
            revert ERC3009ZeroAddress();
        }
        AuthorizationState state = _authorizationStates[authorizer][nonce];
        if (state != AuthorizationState.Unused) {
            revert ERC3009InvalidAuthorization();
        }
        if (block.timestamp <= validAfter) {
            revert ERC3009AuthorizationNotYetValid();
        }
        if (block.timestamp >= validBefore) {
            revert ERC3009AuthorizationExpired();
        }
    }

    function _requireTransferSignature(
        address from,
        address to,
        uint256 value,
        uint256 validAfter,
        uint256 validBefore,
        bytes32 nonce,
        bytes32 typeHash,
        bytes calldata signature
    ) private view {
        bytes32 structHash = keccak256(
            abi.encode(typeHash, from, to, value, validAfter, validBefore, nonce)
        );
        address signer = ECDSA.recoverCalldata(_hashTypedDataV4(structHash), signature);
        if (signer != from) {
            revert ERC3009InvalidSigner();
        }
    }

    function _markAuthorizationUsed(address authorizer, bytes32 nonce) private {
        _authorizationStates[authorizer][nonce] = AuthorizationState.Used;
        emit AuthorizationUsed(authorizer, nonce);
    }
}
