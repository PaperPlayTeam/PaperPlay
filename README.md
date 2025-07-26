# 🎓 PaperPlay - 论文游戏化学习平台

[![Go](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://go.dev)
[![Vue](https://img.shields.io/badge/Vue-2.x-green.svg)](https://vuejs.org)
[![Python](https://img.shields.io/badge/Python-3.10+-orange.svg)](https://python.org)
[![Solidity](https://img.shields.io/badge/Solidity-0.8.28-yellow.svg)](https://soliditylang.org)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

PaperPlay 是一个创新的论文游戏化学习平台，通过AI技术和Web3集成，将复杂的学术论文转化为易于理解的概念和互动式学习体验。

## 🌟 系统架构

项目采用模块化设计，包含四个核心组件：

### 1. 🖥️ 前端 (Vue + uni-app)

- **技术栈**: Vue 2 + uni-app framework
- **跨平台**: 支持H5/小程序/App
- **主要功能**:
  - 用户认证
  - 互动地图
  - 每日挑战
  - 目标追踪
  - 测验系统
  - 个人档案

### 2. ⚡ 后端 (Golang)

- **框架**: Gin + GORM
- **数据库**: SQLite3
- **核心功能**:
  - RESTful API
  - JWT认证
  - WebSocket实时通知
  - 成就系统
  - 性能监控
  - 健康检查

### 3. 🤖 Agent系统 (Python)

- **框架**: LangChain + LangGraph
- **AI模型**: 通义千问
- **主要特性**:
  - PDF智能解析
  - 概念智能抽取
  - 类比式问题生成
  - 混合存储架构
  - 批量处理能力

### 4. 🔗 Web3集成 (Solidity)

- **网络**: Injective EVM
- **合约**: ERC721 NFT
- **功能特点**:
  - 多等级成就
  - NFT徽章
  - 智能合约管理
  - MetaMask集成

## 🚀 快速开始

### 环境要求

- Go 1.21+
- Node.js 16+
- Python 3.10+
- SQLite3
- MetaMask浏览器扩展

### 后端启动

```bash
cd backend
go mod download
go run cmd/main.go
```

### 前端启动

```bash
# 使用HBuilderX（推荐）
# 点击"运行"按钮或使用Ctrl+R

# 或使用CLI
npm run dev:%PLATFORM%     # 开发服务器
npm run build:%PLATFORM%   # 生产构建
```

### Agent系统启动

```bash
cd agent
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
python main.py
```

### Web3部署

```bash
cd injective/hardhat-inj
npm install
npx hardhat compile
npx hardhat run script/deployAchievement.js --network inj_testnet
```

## 📊 系统功能

### 学习功能

- 论文游戏化学习
- 概念可视化
- 互动式测验
- 进度追踪
- 实时反馈

### 成就系统

- 多等级成就
- NFT徽章奖励
- 学习里程碑
- 社区认可

### AI辅助

- 智能概念提取
- 类比式教学
- 个性化推荐
- 学习路径规划

### Web3整合

- 成就通证化
- 链上认证
- 数字资产管理
- 社区激励

## 🔧 项目结构

```
paperplay/
├── backend/           # Golang后端
│   ├── cmd/          # 入口文件
│   ├── config/       # 配置文件
│   ├── internal/     # 内部包
│   └── pkg/          # 公共包
├── frontend/         # Vue前端
│   ├── api/         # API服务
│   ├── pages/       # 页面组件
│   ├── static/      # 静态资源
│   └── utils/       # 工具函数
├── agent/           # AI Agent系统
│   ├── agents/      # Agent模块
│   ├── utils/       # 工具模块
│   └── applications/# 应用模块
└── injective/       # Web3集成
    ├── hardhat-inj/ # 智能合约
    ├── frontend/    # Web3前端
    └── metadata/    # NFT元数据
```

## 📈 性能指标

| 模块            | 指标         | 表现     |
| --------------- | ------------ | -------- |
| **后端**  | API响应时间  | <100ms   |
| **前端**  | 页面加载时间 | <2s      |
| **Agent** | PDF处理速度  | ~30秒/篇 |
| **Web3**  | 交易确认     | <5个区块 |

## 🔮 未来规划

### 短期目标 (1-3个月)

- [ ] 优化用户界面体验
- [ ] 增加更多问题类型
- [ ] 完善概念关联图谱
- [ ] 提升AI处理效率

### 中期目标 (3-6个月)

- [ ] 实现用户个性化推荐
- [ ] 添加多模态支持
- [ ] 扩展Web3功能
- [ ] 优化激励机制

### 长期目标 (6-12个月)

- [ ] 多语言支持
- [ ] 实时协作学习
- [ ] 知识图谱可视化
- [ ] 生态系统扩展

## 👥 贡献指南

欢迎贡献代码！请参考各子目录中的具体贡献指南。

## 🙏 致谢

感谢以下开源项目和技术：

- [LangChain](https://langchain.com)
- [Vue](https://vuejs.org)
- [Gin](https://gin-gonic.com)
- [uni-app](https://uniapp.dcloud.io)
- [OpenZeppelin](https://openzeppelin.com)
- [Injective Protocol](https://injective.com)

---

<div align="center">

**🎓 让学术论文学习变得简单有趣！**

</div>
