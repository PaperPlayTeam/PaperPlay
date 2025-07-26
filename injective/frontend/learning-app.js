// 应用配置
const CONFIG = {
    // 合约地址（部署后需要更新）
    CONTRACT_ADDRESS: '0x0703e331F2E53BFb37461741eDE73eA4d1EEA641', // 学习成就NFT合约地址
    
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
        0: { 
            name: '首战告捷', 
            icon: '🎯',
            description: '第一次作答即答对任意一道题目，获得"首战告捷"成就。',
            condition: '首次答对 ≥ 1 题'
        },
        1: { 
            name: '连击达人', 
            icon: '🔥',
            description: '连续 7 天都有学习记录，保持良好学习习惯，解锁"连击达人"。',
            condition: '连续学习 ≥ 7 天'
        },
        2: { 
            name: '智速双全', 
            icon: '⚡',
            description: '平均每题秒答 ≤30 s 且正确率 ≥ 90%，获得"智速双全"徽章。',
            condition: '平均用时 ≤ 30s 且正确率 ≥ 90%'
        },
        3: { 
            name: '持久战士', 
            icon: '💪',
            description: '单日学习总时长 ≥ 1 小时，解锁"持久战士"勋章。',
            condition: '单日学习时长 ≥ 1 小时'
        },
        4: { 
            name: '零依赖提示', 
            icon: '🧠',
            description: '当日答题 ≥ 20 道且从未使用提示，获得"零依赖提示"称号。',
            condition: '答题 ≥ 20 道且无提示'
        },
        5: { 
            name: '记忆大师', 
            icon: '🎓',
            description: '保留度 ≥ 85% 且当天无需复习推荐，获得"记忆大师"成就。',
            condition: '保留度 ≥ 85% 且无待复习'
        }
    }
};

// 合约ABI
const CONTRACT_ABI = [
    "function name() view returns (string)",
    "function symbol() view returns (string)",
    "function totalSupply() view returns (uint256)",
    "function userLearningData(address user) view returns (tuple(uint256 attempts_first_try_correct, uint256 streak_days, uint256 avg_duration_ms, uint256 correct_rate, uint256 total_time_ms, uint256 hint_used_count, uint256 attempts_total, uint256 retention_score, uint256 review_due_count, uint256 last_update_date))",
    "function getClaimableAchievements(address user) view returns (uint8[])",
    "function getUserAchievements(address user) view returns (uint8[])",
    "function claimAchievement(uint8 achievementType)",
    "function checkAchievementEligibility(address user, uint8 achievementType) view returns (bool)",
    "function getAchievementInfo(uint8 achievementType) view returns (string, string)",
    "function updateUserLearningData(address user, uint256 attempts_first_try_correct, uint256 streak_days, uint256 avg_duration_ms, uint256 correct_rate, uint256 total_time_ms, uint256 hint_used_count, uint256 attempts_total, uint256 retention_score, uint256 review_due_count)",
    "function owner() view returns (address)",
    "event AchievementClaimed(address indexed user, uint256 indexed tokenId, uint8 achievementType, string name)"
];

// 应用状态
class AppState {
    constructor() {
        this.provider = null;
        this.signer = null;
        this.contract = null;
        this.userAddress = null;
        this.isOwner = false;
        this.userLearningData = null;
        this.claimableAchievements = [];
        this.userAchievements = [];
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
    
    // 学习数据显示
    learningDataSection: document.getElementById('learningDataSection'),
    firstTryCorrect: document.getElementById('firstTryCorrect'),
    streakDays: document.getElementById('streakDays'),
    avgDuration: document.getElementById('avgDuration'),
    correctRate: document.getElementById('correctRate'),
    totalTime: document.getElementById('totalTime'),
    hintUsed: document.getElementById('hintUsed'),
    totalAttempts: document.getElementById('totalAttempts'),
    retentionScore: document.getElementById('retentionScore'),
    reviewDue: document.getElementById('reviewDue'),
    
    // 成就相关
    claimableSection: document.getElementById('claimableSection'),
    claimableList: document.getElementById('claimableList'),
    ownedSection: document.getElementById('ownedSection'),
    ownedList: document.getElementById('ownedList'),
    noClaimable: document.getElementById('noClaimable'),
    noOwned: document.getElementById('noOwned'),
    
    // 管理员
    updateDataForm: document.getElementById('updateDataForm')
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

    formatDuration(ms) {
        const seconds = Math.floor(ms / 1000);
        const minutes = Math.floor(seconds / 60);
        const hours = Math.floor(minutes / 60);
        
        if (hours > 0) {
            return `${hours}h ${minutes % 60}m`;
        } else if (minutes > 0) {
            return `${minutes}m ${seconds % 60}s`;
        } else {
            return `${seconds}s`;
        }
    },

    formatPercentage(value) {
        return `${value}%`;
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
            await WalletManager.checkNetwork();

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

            WalletManager.updateUI();
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
        app.userLearningData = null;
        app.claimableAchievements = [];
        app.userAchievements = [];

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

            // 加载用户学习数据
            app.userLearningData = await app.contract.userLearningData(app.userAddress);
            
            // 加载可申领的成就
            app.claimableAchievements = await app.contract.getClaimableAchievements(app.userAddress);
            
            // 加载已拥有的成就
            app.userAchievements = await app.contract.getUserAchievements(app.userAddress);

            // 更新界面显示
            DataManager.updateLearningDataDisplay();
            DataManager.updateAchievementsDisplay();

            utils.hideLoading();
        } catch (error) {
            utils.hideLoading();
            utils.showNotification(`加载数据失败: ${error.message}`, 'error');
        }
    }

    static updateLearningDataDisplay() {
        if (!app.userLearningData) return;

        const data = app.userLearningData;
        
        elements.firstTryCorrect.textContent = data.attempts_first_try_correct.toString();
        elements.streakDays.textContent = data.streak_days.toString();
        elements.avgDuration.textContent = utils.formatDuration(data.avg_duration_ms.toNumber());
        elements.correctRate.textContent = utils.formatPercentage(data.correct_rate.toString());
        elements.totalTime.textContent = utils.formatDuration(data.total_time_ms.toNumber());
        elements.hintUsed.textContent = data.hint_used_count.toString();
        elements.totalAttempts.textContent = data.attempts_total.toString();
        elements.retentionScore.textContent = utils.formatPercentage(data.retention_score.toString());
        elements.reviewDue.textContent = data.review_due_count.toString();
    }

    static updateAchievementsDisplay() {
        // 显示可申领的成就
        if (app.claimableAchievements.length === 0) {
            elements.claimableList.classList.add('hidden');
            elements.noClaimable.classList.remove('hidden');
        } else {
            elements.noClaimable.classList.add('hidden');
            elements.claimableList.classList.remove('hidden');
            
            elements.claimableList.innerHTML = app.claimableAchievements.map(achievementType => {
                const achievement = CONFIG.ACHIEVEMENT_TYPES[achievementType];
                return `
                    <div class="achievement-card claimable">
                        <div class="achievement-icon">${achievement.icon}</div>
                        <div class="achievement-content">
                            <h3>${achievement.name}</h3>
                            <p class="achievement-description">${achievement.description}</p>
                            <p class="achievement-condition">${achievement.condition}</p>
                        </div>
                        <button class="btn btn-primary claim-btn" onclick="AchievementManager.claimAchievement(${achievementType})">
                            申领成就
                        </button>
                    </div>
                `;
            }).join('');
        }

        // 显示已拥有的成就
        if (app.userAchievements.length === 0) {
            elements.ownedList.classList.add('hidden');
            elements.noOwned.classList.remove('hidden');
        } else {
            elements.noOwned.classList.add('hidden');
            elements.ownedList.classList.remove('hidden');
            
            elements.ownedList.innerHTML = app.userAchievements.map(achievementType => {
                const achievement = CONFIG.ACHIEVEMENT_TYPES[achievementType];
                return `
                    <div class="achievement-card owned">
                        <div class="achievement-icon">${achievement.icon}</div>
                        <div class="achievement-content">
                            <h3>${achievement.name}</h3>
                            <p class="achievement-description">${achievement.description}</p>
                            <p class="achievement-condition">${achievement.condition}</p>
                        </div>
                        <div class="achievement-badge">✅ 已获得</div>
                    </div>
                `;
            }).join('');
        }
    }
}

// 成就管理
class AchievementManager {
    static async claimAchievement(achievementType) {
        if (!app.contract) {
            utils.showNotification('合约未连接', 'error');
            return;
        }

        try {
            utils.showLoading();
            utils.showNotification('正在申领成就...', 'info');

            // 估算gas并添加20%余量
            const gasEstimate = await app.contract.estimateGas.claimAchievement(achievementType);
            const gasLimit = gasEstimate.mul(120).div(100); // 增加20%余量

            console.log('预估Gas:', gasEstimate.toString());
            console.log('设置Gas限制:', gasLimit.toString());

            const tx = await app.contract.claimAchievement(achievementType, {
                gasLimit: gasLimit,
                gasPrice: ethers.utils.parseUnits('10', 'gwei') // 10 Gwei
            });
            
            console.log('交易哈希:', tx.hash);
            utils.showNotification(`交易已提交: ${tx.hash.substring(0, 10)}...`, 'info');
            
            const receipt = await tx.wait();
            console.log('交易确认:', receipt);

            utils.showNotification('成就申领成功！', 'success');
            
            // 重新加载数据
            await DataManager.loadUserData();
        } catch (error) {
            console.error('申领错误详情:', error);
            
            let errorMsg = '申领失败';
            if (error.code === 'UNPREDICTABLE_GAS_LIMIT') {
                errorMsg = '交易可能失败，请检查申领条件';
            } else if (error.code === 'INSUFFICIENT_FUNDS') {
                errorMsg = '账户余额不足，请充值INJ代币';
            } else if (error.message.includes('user rejected')) {
                errorMsg = '用户取消了交易';
            } else if (error.message.includes('execution reverted')) {
                errorMsg = '合约执行失败，可能已申领过该成就';
            } else {
                errorMsg = `申领失败: ${error.message}`;
            }
            
            utils.showNotification(errorMsg, 'error');
        } finally {
            utils.hideLoading();
        }
    }
}

// 管理员功能
class AdminManager {
    static async updateUserData(event) {
        event.preventDefault();
        
        if (!app.contract || !app.isOwner) {
            utils.showNotification('需要管理员权限', 'error');
            return;
        }

        const formData = new FormData(event.target);
        const userData = {
            userAddress: formData.get('userAddress'),
            attempts_first_try_correct: parseInt(formData.get('attempts_first_try_correct')),
            streak_days: parseInt(formData.get('streak_days')),
            avg_duration_ms: parseInt(formData.get('avg_duration_ms')),
            correct_rate: parseInt(formData.get('correct_rate')),
            total_time_ms: parseInt(formData.get('total_time_ms')),
            hint_used_count: parseInt(formData.get('hint_used_count')),
            attempts_total: parseInt(formData.get('attempts_total')),
            retention_score: parseInt(formData.get('retention_score')),
            review_due_count: parseInt(formData.get('review_due_count'))
        };

        try {
            utils.showLoading();
            utils.showNotification('正在更新用户数据...', 'info');

            const tx = await app.contract.updateUserLearningData(
                userData.userAddress,
                userData.attempts_first_try_correct,
                userData.streak_days,
                userData.avg_duration_ms,
                userData.correct_rate,
                userData.total_time_ms,
                userData.hint_used_count,
                userData.attempts_total,
                userData.retention_score,
                userData.review_due_count
            );

            await tx.wait();
            utils.showNotification('用户数据更新成功！', 'success');
            
            // 如果更新的是当前用户，重新加载数据
            if (userData.userAddress.toLowerCase() === app.userAddress.toLowerCase()) {
                await DataManager.loadUserData();
            }
        } catch (error) {
            utils.showNotification(`更新失败: ${error.message}`, 'error');
        } finally {
            utils.hideLoading();
        }
    }
}

// 初始化应用
document.addEventListener('DOMContentLoaded', () => {
    // 连接钱包按钮
    elements.connectWallet.addEventListener('click', WalletManager.connect);
    elements.connectPromptBtn.addEventListener('click', WalletManager.connect);
    
    // 断开连接按钮
    elements.disconnectWallet.addEventListener('click', WalletManager.disconnect);
    
    // 管理员表单
    elements.updateDataForm.addEventListener('submit', AdminManager.updateUserData);
    
    // 检查是否已经连接钱包
    if (window.ethereum && window.ethereum.selectedAddress) {
        WalletManager.connect();
    }
    
    // 监听账户变化
    if (window.ethereum) {
        window.ethereum.on('accountsChanged', (accounts) => {
            if (accounts.length === 0) {
                WalletManager.disconnect();
            } else {
                WalletManager.connect();
            }
        });
        
        window.ethereum.on('chainChanged', () => {
            window.location.reload();
        });
    }
}); 