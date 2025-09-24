// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

/**
 * @title SHIBMemeToken
 * @dev SHIB风格Meme代币合约，包含三大核心功能：
 * 1. 交易税机制：每笔交易征收固定比例税费，分配至国库和流动性池
 * 2. 流动性池集成：支持自动添加流动性，关联Uniswap V2风格交易对
 * 3. 交易限制：限制单笔最大交易额度、每日最大交易次数，防止市场操纵
 */
contract SHIBMemeToken is ERC20, Ownable {
    // ========================= 核心配置参数 =========================
    uint256 public constant TOTAL_SUPPLY = 1000000000000 * 10 ** 18; // 总发行量：1万亿枚（10^12）
    uint256 public constant TRANSACTION_TAX_RATE = 5; // 交易税税率：5%
    uint256 public constant TAX_TO_TREASURY_RATIO = 40; // 40%税费进入国库
    uint256 public constant TAX_TO_LIQUIDITY_RATIO = 60; // 60%税费用于流动性
    
    uint256 public constant MAX_TRANSACTION_AMOUNT = 10000000 * 10 ** 18; // 单笔最大交易额度：1000万枚
    uint256 public constant MAX_DAILY_TRANSACTIONS = 10; // 单个地址每日最大交易次数：10次
    
    address public liquidityPoolAddress;
    address public treasuryAddress;
    
    // ========================= 交易限制相关存储 =========================
    mapping(address => mapping(uint256 => uint256)) public dailyTransactionCount;
    mapping(address => bool) public isBlacklisted;

    // ========================= 事件定义 =========================
    event LiquidityPoolUpdated(address indexed newLiquidityPool);
    event TreasuryUpdated(address indexed newTreasury);
    event AccountBlacklisted(address indexed account);
    event AccountWhitelisted(address indexed account);
    event TaxDistributed(address indexed sender, uint256 treasuryAmount, uint256 liquidityAmount);

    // ========================= 构造函数 =========================
    constructor(address _treasuryAddress, address _liquidityPoolAddress) 
        ERC20("SHIB Meme Token", "SHIBMEME") 
        Ownable(msg.sender) 
    {
        require(_treasuryAddress != address(0), "Treasury address cannot be zero");
        require(_liquidityPoolAddress != address(0), "Liquidity pool address cannot be zero");
        
        treasuryAddress = _treasuryAddress;
        liquidityPoolAddress = _liquidityPoolAddress;
        
        _mint(msg.sender, TOTAL_SUPPLY);
    }

    // ========================= 核心功能：交易税机制 =========================
    function _update(address from, address to, uint256 value) internal override {
        // 1. 交易前校验
        _beforeTransferChecks(from, to, value);
        
        // 2. 计算税费
        uint256 taxAmount = 0;
        if (
            from != owner() && 
            to != owner() && 
            from != treasuryAddress && 
            to != treasuryAddress &&
            from != liquidityPoolAddress &&
            to != liquidityPoolAddress
        ) {
            taxAmount = (value * TRANSACTION_TAX_RATE) / 100;
        }
        
        // 3. 分配税费
        uint256 netTransferAmount = value - taxAmount;
        if (taxAmount > 0) {
            uint256 treasuryAmount = (taxAmount * TAX_TO_TREASURY_RATIO) / 100;
            uint256 liquidityAmount = taxAmount - treasuryAmount;
            
            super._update(from, treasuryAddress, treasuryAmount);
            super._update(from, liquidityPoolAddress, liquidityAmount);
            
            emit TaxDistributed(from, treasuryAmount, liquidityAmount);
        }
        
        // 4. 执行核心转账
        super._update(from, to, netTransferAmount);
        
        // 5. 更新交易次数
        if (taxAmount > 0) {
            _updateDailyTransactionCount(from);
        }
    }

    // ========================= 核心功能：交易限制校验 =========================
    function _beforeTransferChecks(address sender, address recipient, uint256 amount) internal view {
        require(!isBlacklisted[sender], "Sender is blacklisted");
        require(!isBlacklisted[recipient], "Recipient is blacklisted");
        
        if (
            sender != owner() && 
            recipient != owner() && 
            sender != treasuryAddress && 
            recipient != treasuryAddress
        ) {
            require(amount <= MAX_TRANSACTION_AMOUNT, "Transaction amount exceeds max limit");
            
            uint256 today = block.timestamp / 86400;
            require(
                dailyTransactionCount[sender][today] < MAX_DAILY_TRANSACTIONS,
                "Daily transaction limit exceeded"
            );
        }
    }

    function _updateDailyTransactionCount(address sender) internal {
        uint256 today = block.timestamp / 86400;
        dailyTransactionCount[sender][today] += 1;
    }

    // ========================= 核心功能：流动性池集成 =========================
    function addLiquidity(uint256 tokenAmount, uint256 ethAmount) external onlyOwner payable {
        require(msg.value == ethAmount, "ETH amount mismatch");
        
        super._update(msg.sender, liquidityPoolAddress, tokenAmount);
        
        (bool success, ) = liquidityPoolAddress.call{value: ethAmount}("");
        require(success, "ETH transfer to liquidity pool failed");
    }

    function removeLiquidity(uint256 liquidityAmount) external onlyOwner {
        (bool success, ) = liquidityPoolAddress.call(
            abi.encodeWithSignature(
                "transferFrom(address,address,uint256)",
                msg.sender,
                address(this),
                liquidityAmount
            )
        );
        require(success, "Liquidity token transfer failed");
        
        (success, ) = liquidityPoolAddress.call(
            abi.encodeWithSignature("burn(address)", msg.sender)
        );
        require(success, "Liquidity removal failed");
    }

    // ========================= 权限管理功能 =========================
    function updateLiquidityPool(address newLiquidityPool) external onlyOwner {
        require(newLiquidityPool != address(0), "New liquidity pool address cannot be zero");
        liquidityPoolAddress = newLiquidityPool;
        emit LiquidityPoolUpdated(newLiquidityPool);
    }

    function updateTreasury(address newTreasury) external onlyOwner {
        require(newTreasury != address(0), "New treasury address cannot be zero");
        treasuryAddress = newTreasury;
        emit TreasuryUpdated(newTreasury);
    }

    function blacklistAccount(address account) external onlyOwner {
        require(account != address(0), "Account cannot be zero");
        require(!isBlacklisted[account], "Account already blacklisted");
        isBlacklisted[account] = true;
        emit AccountBlacklisted(account);
    }

    function whitelistAccount(address account) external onlyOwner {
        require(account != address(0), "Account cannot be zero");
        require(isBlacklisted[account], "Account not blacklisted");
        isBlacklisted[account] = false;
        emit AccountWhitelisted(account);
    }

    // ========================= 兼容ETH接收 =========================
    receive() external payable {}
}