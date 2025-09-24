// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/security/PausableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/token/ERC20/IERC20Upgradeable.sol";
import "@openzeppelin/contracts-upgradeable/token/ERC20/utils/SafeERC20Upgradeable.sol";

contract Stake is Initializable, UUPSUpgradeable, OwnableUpgradeable, PausableUpgradeable {
    using SafeERC20Upgradeable for IERC20Upgradeable;

    // 奖励代币地址 (MetaNode)
    address public rewardToken;
    
    // 每个区块产生的奖励数量
    uint256 public rewardPerBlock;
    
    // 质押池结构体
    struct Pool {
        address stTokenAddress;      // 质押代币地址，address(0)表示原生代币
        uint256 poolWeight;          // 池权重，影响奖励分配
        uint256 lastRewardBlock;     // 最后一次计算奖励的区块号
        uint256 accMetaNodePerST;    // 每个质押代币累积的奖励数量
        uint256 stTokenAmount;       // 池中的总质押代币量
        uint256 minDepositAmount;    // 最小质押金额
        uint256 unstakeLockedBlocks; // 解除质押的锁定区块数
    }
    
    // 用户结构体
    struct User {
        uint256 stAmount;            // 用户质押的代币数量
        uint256 finishedMetaNode;    // 已分配的奖励数量
        uint256 pendingMetaNode;     // 待领取的奖励数量
        UnstakeRequest[] requests;   // 解质押请求列表
    }
    
    // 解质押请求结构体
    struct UnstakeRequest {
        uint256 amount;              // 解质押数量
        uint256 unlockBlock;         // 解锁区块号
    }
    
    // 质押池列表
    Pool[] public pools;
    
    // 用户数据映射: poolId => userAddress => User
    mapping(uint256 => mapping(address => User)) public users;
    
    // 功能暂停状态
    mapping(string => bool) public pausedFunctions;
    
    // 事件定义
    event Staked(address indexed user, uint256 indexed pid, uint256 amount);
    event UnstakeRequested(address indexed user, uint256 indexed pid, uint256 amount, uint256 unlockBlock);
    event UnstakeClaimed(address indexed user, uint256 indexed pid, uint256 amount);
    event RewardClaimed(address indexed user, uint256 indexed pid, uint256 amount);
    event PoolAdded(uint256 indexed pid, address stToken, uint256 weight);
    event PoolUpdated(uint256 indexed pid, address stToken, uint256 weight);
    event FunctionPaused(string functionName, bool paused);
    event RewardPerBlockUpdated(uint256 oldValue, uint256 newValue);
    
    // 初始化函数
    function initialize(address _rewardToken, uint256 _rewardPerBlock) external initializer {
        __Ownable_init();
        __Pausable_init();
        __UUPSUpgradeable_init();
        
        rewardToken = _rewardToken;
        rewardPerBlock = _rewardPerBlock;
    }
    
    // 权限控制: 只有管理员可以调用
    modifier onlyAdmin() {
        require(msg.sender == owner(), "Stake: not admin");
        _;
    }
    
    // 功能暂停控制
    modifier whenFunctionNotPaused(string memory functionName) {
        require(!pausedFunctions[functionName], "Stake: function paused");
        _;
    }
    
    // 计算奖励
    function updatePool(uint256 _pid) public {
        Pool storage pool = pools[_pid];
        if (block.number <= pool.lastRewardBlock) {
            return;
        }
        
        uint256 totalWeight = 0;
        for (uint256 i = 0; i < pools.length; i++) {
            totalWeight += pools[i].poolWeight;
        }
        
        if (totalWeight == 0) {
            pool.lastRewardBlock = block.number;
            return;
        }
        
        uint256 blockCount = block.number - pool.lastRewardBlock;
        uint256 reward = (blockCount * rewardPerBlock * pool.poolWeight) / totalWeight;
        
        if (pool.stTokenAmount > 0) {
            pool.accMetaNodePerST += (reward * 1e18) / pool.stTokenAmount;
        }
        
        pool.lastRewardBlock = block.number;
    }
    
    // 质押
    function stake(uint256 _pid, uint256 _amount) external payable whenNotPaused whenFunctionNotPaused("stake") {
        require(_pid < pools.length, "Stake: invalid pool ID");
        Pool storage pool = pools[_pid];
        require(_amount >= pool.minDepositAmount, "Stake: amount below minimum");
        
        User storage user = users[_pid][msg.sender];
        
        // 更新奖励
        updatePool(_pid);
        if (user.stAmount > 0) {
            uint256 pending = (user.stAmount * pool.accMetaNodePerST) / 1e18 - user.finishedMetaNode;
            if (pending > 0) {
                user.pendingMetaNode += pending;
            }
        }
        
        // 处理质押
        if (pool.stTokenAddress == address(0)) {
            // 原生代币质押
            require(msg.value == _amount, "Stake: incorrect ETH amount");
        } else {
            // ERC20代币质押
            IERC20Upgradeable(pool.stTokenAddress).safeTransferFrom(msg.sender, address(this), _amount);
        }
        
        // 更新用户和池数据
        user.stAmount += _amount;
        pool.stTokenAmount += _amount;
        user.finishedMetaNode = (user.stAmount * pool.accMetaNodePerST) / 1e18;
        
        emit Staked(msg.sender, _pid, _amount);
    }
    
    // 请求解除质押
    function requestUnstake(uint256 _pid, uint256 _amount) external whenNotPaused whenFunctionNotPaused("unstake") {
        require(_pid < pools.length, "Stake: invalid pool ID");
        Pool storage pool = pools[_pid];
        User storage user = users[_pid][msg.sender];
        
        require(user.stAmount >= _amount, "Stake: insufficient staked amount");
        
        // 更新奖励
        updatePool(_pid);
        uint256 pending = (user.stAmount * pool.accMetaNodePerST) / 1e18 - user.finishedMetaNode;
        if (pending > 0) {
            user.pendingMetaNode += pending;
        }
        
        // 创建解质押请求
        uint256 unlockBlock = block.number + pool.unstakeLockedBlocks;
        user.requests.push(UnstakeRequest({
            amount: _amount,
            unlockBlock: unlockBlock
        }));
        
        // 更新用户和池数据
        user.stAmount -= _amount;
        pool.stTokenAmount -= _amount;
        user.finishedMetaNode = (user.stAmount * pool.accMetaNodePerST) / 1e18;
        
        emit UnstakeRequested(msg.sender, _pid, _amount, unlockBlock);
    }
    
    // 领取已解锁的解除质押代币
    function claimUnstaked(uint256 _pid) external whenNotPaused whenFunctionNotPaused("claimUnstaked") {
        require(_pid < pools.length, "Stake: invalid pool ID");
        Pool storage pool = pools[_pid];
        User storage user = users[_pid][msg.sender];
        
        uint256 totalClaimable = 0;
        uint256 i = 0;
        
        // 收集所有已解锁的请求
        while (i < user.requests.length) {
            if (block.number >= user.requests[i].unlockBlock) {
                totalClaimable += user.requests[i].amount;
                
                // 移除已处理的请求（与最后一个元素交换并弹出）
                if (i < user.requests.length - 1) {
                    user.requests[i] = user.requests[user.requests.length - 1];
                }
                user.requests.pop();
            } else {
                i++;
            }
        }
        
        require(totalClaimable > 0, "Stake: no claimable unstaked amount");
        
        // 转移代币给用户
        if (pool.stTokenAddress == address(0)) {
            // 原生代币
            (bool success, ) = msg.sender.call{value: totalClaimable}("");
            require(success, "Stake: ETH transfer failed");
        } else {
            // ERC20代币
            IERC20Upgradeable(pool.stTokenAddress).safeTransfer(msg.sender, totalClaimable);
        }
        
        emit UnstakeClaimed(msg.sender, _pid, totalClaimable);
    }
    
    // 领取奖励
    function claimReward(uint256 _pid) external whenNotPaused whenFunctionNotPaused("claimReward") {
        require(_pid < pools.length, "Stake: invalid pool ID");
        Pool storage pool = pools[_pid];
        User storage user = users[_pid][msg.sender];
        
        // 更新奖励
        updatePool(_pid);
        uint256 pending = (user.stAmount * pool.accMetaNodePerST) / 1e18 - user.finishedMetaNode + user.pendingMetaNode;
        require(pending > 0, "Stake: no reward to claim");
        
        // 重置用户奖励数据
        user.finishedMetaNode = (user.stAmount * pool.accMetaNodePerST) / 1e18;
        user.pendingMetaNode = 0;
        
        // 转移奖励代币
        IERC20Upgradeable(rewardToken).safeTransfer(msg.sender, pending);
        
        emit RewardClaimed(msg.sender, _pid, pending);
    }
    
    // 添加质押池
    function addPool(
        address _stTokenAddress,
        uint256 _poolWeight,
        uint256 _minDepositAmount,
        uint256 _unstakeLockedBlocks
    ) external onlyAdmin {
        require(_poolWeight > 0, "Stake: weight must be positive");
        require(_minDepositAmount >= 0, "Stake: min deposit cannot be negative");
        require(_unstakeLockedBlocks >= 0, "Stake: lock blocks cannot be negative");
        
        uint256 pid = pools.length;
        pools.push(Pool({
            stTokenAddress: _stTokenAddress,
            poolWeight: _poolWeight,
            lastRewardBlock: block.number,
            accMetaNodePerST: 0,
            stTokenAmount: 0,
            minDepositAmount: _minDepositAmount,
            unstakeLockedBlocks: _unstakeLockedBlocks
        }));
        
        emit PoolAdded(pid, _stTokenAddress, _poolWeight);
    }
    
    // 更新质押池
    function updatePool(
        uint256 _pid,
        uint256 _poolWeight,
        uint256 _minDepositAmount,
        uint256 _unstakeLockedBlocks
    ) external onlyAdmin {
        require(_pid < pools.length, "Stake: invalid pool ID");
        require(_poolWeight > 0, "Stake: weight must be positive");
        require(_minDepositAmount >= 0, "Stake: min deposit cannot be negative");
        require(_unstakeLockedBlocks >= 0, "Stake: lock blocks cannot be negative");
        
        // 先更新奖励计算
        updatePool(_pid);
        
        Pool storage pool = pools[_pid];
        pool.poolWeight = _poolWeight;
        pool.minDepositAmount = _minDepositAmount;
        pool.unstakeLockedBlocks = _unstakeLockedBlocks;
        
        emit PoolUpdated(_pid, pool.stTokenAddress, _poolWeight);
    }
    
    // 暂停/恢复功能
    function setFunctionPause(string calldata _functionName, bool _paused) external onlyAdmin {
        pausedFunctions[_functionName] = _paused;
        emit FunctionPaused(_functionName, _paused);
    }
    
    // 更新每个区块的奖励数量
    function setRewardPerBlock(uint256 _newRewardPerBlock) external onlyAdmin {
        emit RewardPerBlockUpdated(rewardPerBlock, _newRewardPerBlock);
        rewardPerBlock = _newRewardPerBlock;
    }
    
    // 获取用户的解质押请求
    function getUserUnstakeRequests(uint256 _pid, address _user) external view returns (UnstakeRequest[] memory) {
        return users[_pid][_user].requests;
    }
    
    // 获取质押池数量
    function poolLength() external view returns (uint256) {
        return pools.length;
    }
    
    // 计算用户可领取的奖励
    function pendingReward(uint256 _pid, address _user) external view returns (uint256) {
        Pool storage pool = pools[_pid];
        User storage user = users[_pid][_user];
        
        uint256 accMetaNodePerST = pool.accMetaNodePerST;
        uint256 stTokenAmount = pool.stTokenAmount;
        
        if (block.number > pool.lastRewardBlock && stTokenAmount > 0) {
            uint256 totalWeight = 0;
            for (uint256 i = 0; i < pools.length; i++) {
                totalWeight += pools[i].poolWeight;
            }
            
            uint256 blockCount = block.number - pool.lastRewardBlock;
            uint256 reward = (blockCount * rewardPerBlock * pool.poolWeight) / totalWeight;
            accMetaNodePerST += (reward * 1e18) / stTokenAmount;
        }
        
        return (user.stAmount * accMetaNodePerST) / 1e18 - user.finishedMetaNode + user.pendingMetaNode;
    }
    
    // 升级权限控制
    function _authorizeUpgrade(address newImplementation) internal override onlyAdmin {}
    
    // 接收原生代币
    receive() external payable {}
}
