# 🏆 NFT成就系统

一个基于Injective EVM的极简NFT成就系统，支持多等级成就徽章的铸造和管理。

## ✨ 系统特性

### 🎯 核心功能
- **多等级成就**: 支持青铜🥉、白银🥈、黄金🥇、钻石💎四个等级
- **智能合约**: 基于ERC721标准的NFT成就合约
- **Web界面**: 现代化的前端界面，支持MetaMask连接
- **管理员功能**: 成就铸造、批量操作等管理功能
- **用户统计**: 实时统计各等级成就数量
- **元数据支持**: 完整的NFT元数据体系

### 🛠 技术栈
- **智能合约**: Solidity 0.8.28 + OpenZeppelin
- **开发框架**: Hardhat
- **前端**: HTML5 + CSS3 + JavaScript ES6
- **钱包集成**: MetaMask + Ethers.js
- **网络**: Injective EVM Testnet

## 📁 项目结构

```
injective/
├── hardhat-inj/                 # 智能合约项目
│   ├── contracts/
│   │   ├── AchievementNFT.sol   # 主要NFT成就合约
│   │   └── Counter.sol          # 示例合约
│   ├── script/
│   │   ├── deploy.js            # 原始部署脚本
│   │   └── deployAchievement.js # 成就合约部署脚本
│   ├── test/
│   │   ├── Counter.test.js      # 原始测试
│   │   └── AchievementNFT.test.js # 成就合约测试
│   ├── package.json             # 项目依赖
│   └── hardhat.config.js        # Hardhat配置
├── frontend/                    # 前端界面
│   ├── index.html              # 主页面
│   ├── style.css               # 样式文件
│   └── app.js                  # 应用逻辑
├── metadata/                   # NFT元数据
│   ├── bronze.json             # 青铜成就元数据
│   ├── silver.json             # 白银成就元数据
│   ├── gold.json               # 黄金成就元数据
│   └── diamond.json            # 钻石成就元数据
└── README.md                   # 项目说明
```

## 🚀 快速开始

### 1. 环境准备

确保您已安装以下工具：
- [Node.js](https://nodejs.org/) (v16+)
- [MetaMask](https://metamask.io/) 浏览器扩展
- Git

### 2. 安装依赖

```bash
cd hardhat-inj
npm install
```

### 3. 环境配置

创建 `.env` 文件：
```bash
cp .env.example .env
```

配置环境变量：
```env
PRIVATE_KEY=你的私钥
INJ_TESTNET_RPC_URL=https://k8s.testnet.json-rpc.injective.network/
```

### 4. 编译合约

```bash
npx hardhat compile
```

### 5. 运行测试

```bash
npx hardhat test
```

### 6. 部署合约

```bash
# 部署到Injective测试网
npx hardhat run script/deployAchievement.js --network inj_testnet
```

部署成功后，会生成 `deployment-info.json` 文件，包含合约地址等信息。

### 7. 配置前端

编辑 `frontend/app.js`，更新合约地址：
```javascript
const CONFIG = {
    CONTRACT_ADDRESS: '0x您的合约地址', // 替换为实际部署的地址
    // ... 其他配置
};
```

### 8. 启动前端

```bash
# 方法1: 使用简单HTTP服务器
cd frontend
python -m http.server 8000

# 方法2: 使用Node.js服务器
npx http-server frontend -p 8000

# 方法3: 使用Live Server (VS Code扩展)
# 右键index.html -> Open with Live Server
```

访问 http://localhost:8000 查看应用。

## 📖 使用指南

### 🔐 连接钱包

1. 确保MetaMask已安装并连接到Injective测试网
2. 点击"连接钱包"按钮
3. 授权MetaMask访问

### 🎖️ 查看成就

连接钱包后，您可以：
- 查看成就统计（总数、各等级数量）
- 浏览已获得的成就列表
- 查看成就详细信息

### ⚙️ 管理员功能

如果您是合约所有者，可以：

#### 铸造单个成就
1. 在管理员面板填写表单：
   - 接收者地址
   - 成就名称
   - 成就描述
   - 成就等级
2. 点击"铸造成就"提交交易

#### 批量铸造成就
```javascript
// 在浏览器控制台执行
const recipients = ['0x地址1', '0x地址2', '0x地址3'];
await app.contract.batchMintAchievement(
    recipients,
    "活动参与者",
    "感谢参与我们的特殊活动！",
    0 // 青铜等级
);
```

## 🎨 成就系统设计

### 等级体系
| 等级 | 图标 | 分数 | 稀有度 | 用途 |
|------|------|------|--------|------|
| 青铜 | 🥉 | 100 | 常见 | 基础成就、参与奖励 |
| 白银 | 🥈 | 250 | 不常见 | 进阶成就、活跃用户 |
| 黄金 | 🥇 | 500 | 稀有 | 高级成就、重要里程碑 |
| 钻石 | 💎 | 1000 | 传奇 | 最高荣誉、特殊贡献 |

### 应用场景
- **学习平台**: 课程完成、技能掌握
- **游戏应用**: 任务完成、排行榜奖励
- **社区活动**: 参与奖励、贡献认可
- **企业培训**: 培训完成、认证获得

## 🔧 开发扩展

### 添加新功能

#### 1. 成就转移功能
```solidity
function transferAchievement(address from, address to, uint256 tokenId) external onlyOwner {
    _transfer(from, to, tokenId);
}
```

#### 2. 成就升级机制
```solidity
function upgradeAchievement(uint256 tokenId, AchievementType newType) external onlyOwner {
    // 升级逻辑
}
```

#### 3. 成就销毁功能
```solidity
function burnAchievement(uint256 tokenId) external {
    require(ownerOf(tokenId) == msg.sender || msg.sender == owner(), "Not authorized");
    _burn(tokenId);
}
```

### 前端自定义

#### 修改主题色彩
在 `frontend/style.css` 中：
```css
:root {
    --primary-color: #007bff;
    --secondary-color: #6c757d;
    --success-color: #28a745;
    --warning-color: #ffc107;
    --error-color: #dc3545;
}
```

#### 添加新的成就类型
1. 修改合约的 `AchievementType` 枚举
2. 更新前端 `CONFIG.ACHIEVEMENT_TYPES`
3. 创建对应的元数据文件

## 🧪 测试指南

### 运行完整测试套件
```bash
npx hardhat test
```

### 运行特定测试
```bash
npx hardhat test test/AchievementNFT.test.js
```

### 测试覆盖率
```bash
npx hardhat coverage
```

## 🌐 网络配置

### Injective Testnet
- **Chain ID**: 1439 (0x59F)
- **RPC URL**: https://k8s.testnet.json-rpc.injective.network/
- **浏览器**: https://testnet.blockscout.injective.network/
- **水龙头**: [Injective Faucet](https://testnet.faucet.injective.network/)

### 切换到主网
修改 `hardhat.config.js`：
```javascript
inj_mainnet: {
    url: 'https://k8s.mainnet.json-rpc.injective.network/',
    accounts: process.env.PRIVATE_KEY ? [process.env.PRIVATE_KEY] : [],
    chainId: 1234, // Injective主网Chain ID
}
```

## 📝 智能合约API

### 主要函数

#### 查询功能
```solidity
// 获取用户总成就数
getUserTotalAchievements(address user) → uint256

// 获取特定等级成就数
getUserAchievementsByType(address user, AchievementType type) → uint256

// 检查是否拥有特定成就
hasAchievement(address user, AchievementType type) → bool

// 获取成就详情
getAchievement(uint256 tokenId) → Achievement
```

#### 管理功能（仅所有者）
```solidity
// 铸造成就
mintAchievement(address to, string name, string description, AchievementType type) → uint256

// 批量铸造
batchMintAchievement(address[] recipients, string name, string description, AchievementType type)

// 设置基础URI
setBaseTokenURI(AchievementType type, string baseURI)
```

## 🔒 安全注意事项

### 智能合约安全
- ✅ 使用OpenZeppelin经过审计的库
- ✅ 实现访问控制（onlyOwner）
- ✅ 输入验证和错误处理
- ✅ 重入攻击防护

### 前端安全
- ✅ 钱包连接验证
- ✅ 交易确认机制
- ✅ 网络检查
- ✅ 地址格式验证

### 部署安全
- 🔐 私钥安全管理
- 🔐 环境变量保护
- 🔐 合约验证和开源

## 🤝 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/新功能`)
3. 提交更改 (`git commit -m '添加新功能'`)
4. 推送到分支 (`git push origin feature/新功能`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🆘 故障排除

### 常见问题

#### Q: 合约部署失败
A: 检查：
- 私钥是否正确配置
- 账户是否有足够的INJ代币
- 网络连接是否正常

#### Q: 前端无法连接合约
A: 确认：
- 合约地址是否正确更新
- MetaMask是否连接到正确网络
- 浏览器控制台是否有错误信息

#### Q: 交易失败
A: 检查：
- Gas费用是否足够
- 合约权限是否正确
- 参数格式是否正确

### 获取帮助
- 📧 邮箱: support@achievementnft.com
- 💬 Discord: [加入我们的社区](https://discord.gg/achievementnft)
- 🐛 问题反馈: [GitHub Issues](https://github.com/your-repo/issues)

---

**🎉 开始您的NFT成就之旅吧！** 