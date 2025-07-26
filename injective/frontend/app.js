// 应用配置
const CONFIG = {
    // 合约地址（已部署）
    CONTRACT_ADDRESS: '0x0703e331F2E53BFb37461741eDE73eA4d1EEA641', // AchievementNFT合约地址
    
    // 网络配置
    NETWORK: {
        chainId: '0x59F', // 1439 in hex (Injective testnet)
        chainName: 'Injective Testnet',
        rpcUrls: ['https://k8s.testnet.json-rpc.injective.network/'],
        nativeCurrency: {
            name: 'INJ',
            symbol: 'INJ',
            decimals: 18
        },
        blockExplorerUrls: ['https://testnet.blockscout.injective.network/']
    },
    
    // 成就类型映射
    ACHIEVEMENT_TYPES: {
        0: { name: '青铜', icon: '🥉', class: 'bronze' },
        1: { name: '白银', icon: '🥈', class: 'silver' },
        2: { name: '黄金', icon: '🥇', class: 'gold' },
        3: { name: '钻石', icon: '💎', class: 'diamond' }
    }
};

// 合约ABI（简化版）
const CONTRACT_ABI = [
    "function name() view returns (string)",
    "function symbol() view returns (string)",
    "function totalSupply() view returns (uint256)",
    "function getUserTotalAchievements(address user) view returns (uint256)",
    "function getUserAchievementsByType(address user, uint8 achievementType) view returns (uint256)",
    "function hasAchievement(address user, uint8 achievementType) view returns (bool)",
    "function getAchievement(uint256 tokenId) view returns (tuple(string name, string description, uint8 achievementType, uint256 timestamp, address recipient))",
    "function mintAchievement(address to, string name, string description, uint8 achievementType) returns (uint256)",
    "function batchMintAchievement(address[] recipients, string name, string description, uint8 achievementType)",
    "function owner() view returns (address)",
    "function balanceOf(address owner) view returns (uint256)",
    "function tokenOfOwnerByIndex(address owner, uint256 index) view returns (uint256)",
    "event AchievementMinted(address indexed recipient, uint256 indexed tokenId, uint8 achievementType, string name)"
];

// 应用状态
class AppState {
    constructor() {
        this.provider = null;
        this.signer = null;
        this.contract = null;
        this.userAddress = null;
        this.isOwner = false;
        this.achievements = [];
    }
}

const app = new AppState();

// DOM 元素
const elements = {
    connectWallet: document.getElementById('connectWallet'),
    connectPromptBtn: document.getElementById('connectPromptBtn'),
    disconnectWallet: document.getElementById('disconnectWallet'),
    walletInfo: document.getElementById('walletInfo'),
    walletAddress: document.getElementById('walletAddress'),
    connectPrompt: document.getElementById('connectPrompt'),
    mainInterface: document.getElementById('mainInterface'),
    adminSection: document.getElementById('adminSection'),
    loadingOverlay: document.getElementById('loadingOverlay'),
    notifications: document.getElementById('notifications'),
    mintForm: document.getElementById('mintForm'),
    achievementsList: document.getElementById('achievementsList'),
    noAchievements: document.getElementById('noAchievements'),
    totalAchievements: document.getElementById('totalAchievements'),
    bronzeCount: document.getElementById('bronzeCount'),
    silverCount: document.getElementById('silverCount'),
    goldCount: document.getElementById('goldCount'),
    diamondCount: document.getElementById('diamondCount')
};

// 工具函数
const utils = {
    showLoading() {
        elements.loadingOverlay.classList.remove('hidden');
    },

    hideLoading() {
        elements.loadingOverlay.classList.add('hidden');
    },

    showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.innerHTML = `
            <div style="font-weight: 600; margin-bottom: 4px;">${type === 'success' ? '✅' : type === 'error' ? '❌' : 'ℹ️'} ${type.toUpperCase()}</div>
            <div>${message}</div>
        `;
        
        elements.notifications.appendChild(notification);
        
        setTimeout(() => {
            notification.remove();
        }, 5000);
    },

    formatAddress(address) {
        return `${address.slice(0, 6)}...${address.slice(-4)}`;
    },

    formatTimestamp(timestamp) {
        return new Date(timestamp * 1000).toLocaleDateString('zh-CN');
    }
};

// 钱包连接功能
class WalletManager {
    static async connect() {
        try {
            if (!window.ethereum) {
                utils.showNotification('请安装 MetaMask 钱包', 'error');
                return false;
            }

            utils.showLoading();

            // 请求账户访问
            const accounts = await window.ethereum.request({
                method: 'eth_requestAccounts'
            });

            if (accounts.length === 0) {
                utils.showNotification('未找到账户', 'error');
                return false;
            }

            // 检查网络
            //await this.checkNetwork();

            // 设置提供者和签名者
            app.provider = new ethers.providers.Web3Provider(window.ethereum);
            app.signer = app.provider.getSigner();
            app.userAddress = accounts[0];

            // 初始化合约
            if (CONFIG.CONTRACT_ADDRESS && CONFIG.CONTRACT_ADDRESS !== '0x...') {
                app.contract = new ethers.Contract(CONFIG.CONTRACT_ADDRESS, CONTRACT_ABI, app.signer);
                
                // 检查是否为合约所有者
                try {
                    const owner = await app.contract.owner();
                    app.isOwner = owner.toLowerCase() === app.userAddress.toLowerCase();
                } catch (error) {
                    console.warn('无法检查合约所有者权限:', error);
                }
            }

            this.updateUI();
            utils.hideLoading();
            utils.showNotification('钱包连接成功！', 'success');

            // 如果合约已部署，加载数据
            if (app.contract) {
                await DataManager.loadUserData();
            }

            return true;
        } catch (error) {
            utils.hideLoading();
            utils.showNotification(`连接失败: ${error.message}`, 'error');
            return false;
        }
    }

    static async checkNetwork() {
        const chainId = await window.ethereum.request({ method: 'eth_chainId' });
        
        if (chainId !== CONFIG.NETWORK.chainId) {
            try {
                await window.ethereum.request({
                    method: 'wallet_switchEthereumChain',
                    params: [{ chainId: CONFIG.NETWORK.chainId }]
                });
            } catch (switchError) {
                if (switchError.code === 4902) {
                    await window.ethereum.request({
                        method: 'wallet_addEthereumChain',
                        params: [CONFIG.NETWORK]
                    });
                } else {
                    throw switchError;
                }
            }
        }
    }

    static disconnect() {
        app.provider = null;
        app.signer = null;
        app.contract = null;
        app.userAddress = null;
        app.isOwner = false;
        app.achievements = [];

        elements.connectPrompt.classList.remove('hidden');
        elements.mainInterface.classList.add('hidden');
        elements.walletInfo.classList.add('hidden');
        elements.connectWallet.classList.remove('hidden');

        utils.showNotification('钱包已断开连接', 'info');
    }

    static updateUI() {
        if (app.userAddress) {
            elements.walletAddress.textContent = utils.formatAddress(app.userAddress);
            elements.walletInfo.classList.remove('hidden');
            elements.connectWallet.classList.add('hidden');
            elements.connectPrompt.classList.add('hidden');
            elements.mainInterface.classList.remove('hidden');

            // 显示管理员面板
            if (app.isOwner) {
                elements.adminSection.classList.remove('hidden');
            }
        }
    }
}

// 数据管理
class DataManager {
    static async loadUserData() {
        if (!app.contract || !app.userAddress) {
            return;
        }

        try {
            utils.showLoading();

            // 加载统计数据
            const [total, bronze, silver, gold, diamond] = await Promise.all([
                app.contract.getUserTotalAchievements(app.userAddress),
                app.contract.getUserAchievementsByType(app.userAddress, 0),
                app.contract.getUserAchievementsByType(app.userAddress, 1),
                app.contract.getUserAchievementsByType(app.userAddress, 2),
                app.contract.getUserAchievementsByType(app.userAddress, 3)
            ]);

            // 更新统计显示
            elements.totalAchievements.textContent = total.toString();
            elements.bronzeCount.textContent = bronze.toString();
            elements.silverCount.textContent = silver.toString();
            elements.goldCount.textContent = gold.toString();
            elements.diamondCount.textContent = diamond.toString();

            // 加载用户的成就列表
            await this.loadUserAchievements();

            utils.hideLoading();
        } catch (error) {
            utils.hideLoading();
            utils.showNotification(`加载数据失败: ${error.message}`, 'error');
        }
    }

    static async loadUserAchievements() {
        if (!app.contract || !app.userAddress) {
            return;
        }

        try {
            // 获取用户拥有的NFT数量
            const balance = await app.contract.balanceOf(app.userAddress);
            const achievements = [];

            // 由于没有tokenOfOwnerByIndex函数，我们需要监听事件来获取用户的成就
            // 这里简化处理，实际应用中应该使用事件过滤或其他方法
            
            this.displayAchievements(achievements);
        } catch (error) {
            console.error('加载用户成就失败:', error);
        }
    }

    static displayAchievements(achievements) {
        if (achievements.length === 0) {
            elements.achievementsList.classList.add('hidden');
            elements.noAchievements.classList.remove('hidden');
            return;
        }

        elements.noAchievements.classList.add('hidden');
        elements.achievementsList.classList.remove('hidden');

        elements.achievementsList.innerHTML = achievements.map(achievement => `
            <div class="achievement-card">
                <div class="achievement-header">
                    <div class="achievement-icon">${CONFIG.ACHIEVEMENT_TYPES[achievement.achievementType].icon}</div>
                    <div class="achievement-info">
                        <h3>${achievement.name}</h3>
                        <span class="achievement-type ${CONFIG.ACHIEVEMENT_TYPES[achievement.achievementType].class}">
                            ${CONFIG.ACHIEVEMENT_TYPES[achievement.achievementType].name}
                        </span>
                    </div>
                </div>
                <div class="achievement-description">${achievement.description}</div>
                <div class="achievement-meta">
                    <span>获得时间: ${utils.formatTimestamp(achievement.timestamp)}</span>
                    <span>Token ID: #${achievement.tokenId}</span>
                </div>
            </div>
        `).join('');
    }
}

// 管理员功能
class AdminManager {
    static async mintAchievement(recipient, name, description, type) {
        if (!app.contract || !app.isOwner) {
            utils.showNotification('只有合约所有者可以铸造成就', 'error');
            return;
        }

        try {
            utils.showLoading();

            const tx = await app.contract.mintAchievement(recipient, name, description, type);
            utils.showNotification('交易已提交，等待确认...', 'info');

            const receipt = await tx.wait();
            
            utils.showNotification(`成就铸造成功！交易哈希: ${receipt.transactionHash}`, 'success');
            
            // 刷新数据
            await DataManager.loadUserData();
            
        } catch (error) {
            utils.showNotification(`铸造失败: ${error.message}`, 'error');
        } finally {
            utils.hideLoading();
        }
    }
}

// 事件监听器
function setupEventListeners() {
    // 钱包连接
    elements.connectWallet.addEventListener('click', WalletManager.connect);
    elements.connectPromptBtn.addEventListener('click', WalletManager.connect);
    elements.disconnectWallet.addEventListener('click', WalletManager.disconnect);

    // 管理员表单
    elements.mintForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const recipient = document.getElementById('recipientAddress').value;
        const name = document.getElementById('achievementName').value;
        const description = document.getElementById('achievementDescription').value;
        const type = parseInt(document.getElementById('achievementType').value);

        if (!ethers.utils.isAddress(recipient)) {
            utils.showNotification('无效的地址格式', 'error');
            return;
        }

        await AdminManager.mintAchievement(recipient, name, description, type);
        
        // 清空表单
        elements.mintForm.reset();
    });

    // 监听账户变化
    if (window.ethereum) {
        window.ethereum.on('accountsChanged', (accounts) => {
            if (accounts.length === 0) {
                WalletManager.disconnect();
            } else {
                location.reload();
            }
        });

        window.ethereum.on('chainChanged', () => {
            location.reload();
        });
    }
}

// 配置更新函数（部署后调用）
function updateContractAddress(address) {
    CONFIG.CONTRACT_ADDRESS = address;
    if (app.signer) {
        app.contract = new ethers.Contract(address, CONTRACT_ABI, app.signer);
    }
}

// 应用初始化
function init() {
    setupEventListeners();
    
    // 检查是否已连接钱包
    if (window.ethereum && window.ethereum.selectedAddress) {
        WalletManager.connect();
    }

    // 检查合约地址是否已配置
    if (CONFIG.CONTRACT_ADDRESS === '0x...') {
        utils.showNotification('合约地址未配置，请先部署合约并更新配置', 'warning');
    }

    console.log('🏆 NFT成就系统已初始化');
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', init); 