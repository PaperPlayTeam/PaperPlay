import { ethers } from 'ethers';

// 网络配置
export const NETWORK_CONFIG = {
    chainId: '0x59F', // 1439 in hex (Injective testnet)
    chainName: 'Injective Testnet',
    rpcUrls: ['https://k8s.testnet.json-rpc.injective.network/'],
    nativeCurrency: {
        name: 'INJ',
        symbol: 'INJ',
        decimals: 18
    },
    blockExplorerUrls: ['https://testnet.blockscout.injective.network/']
};

// 成就类型映射
export const ACHIEVEMENT_TYPES = {
    0: { name: '青铜', icon: '🥉', class: 'bronze' },
    1: { name: '白银', icon: '🥈', class: 'silver' },
    2: { name: '黄金', icon: '🥇', class: 'gold' },
    3: { name: '钻石', icon: '💎', class: 'diamond' }
};

// 合约配置
export const CONTRACT_CONFIG = {
    address: '0x0703e331F2E53BFb37461741eDE73eA4d1EEA641', // AchievementNFT合约地址
    abi: [
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
    ]
};

let provider = null;
let signer = null;

// Initialize Web3
export const initWeb3 = async () => {
    try {
        // Check if we're in browser environment and ethereum is available
        if (typeof window !== 'undefined' && window.ethereum) {
            // Request account access
            await window.ethereum.request({ method: 'eth_requestAccounts' });
            
            // Create provider
            provider = new ethers.providers.Web3Provider(window.ethereum);
            signer = provider.getSigner();
            
            return { provider, signer };
        } else {
            throw new Error('请在支持injective测试网的环境中运行');
        }
    } catch (error) {
        console.error('初始化Web3失败:', error);
        throw error;
    }
};

// 检查网络
async function checkNetwork() {
    const chainId = await window.ethereum.request({ method: 'eth_chainId' });
    
    if (chainId !== NETWORK_CONFIG.chainId) {
        try {
            await window.ethereum.request({
                method: 'wallet_switchEthereumChain',
                params: [{ chainId: NETWORK_CONFIG.chainId }]
            });
        } catch (switchError) {
            // 如果网络不存在，添加网络
            if (switchError.code === 4902) {
                await window.ethereum.request({
                    method: 'wallet_addEthereumChain',
                    params: [NETWORK_CONFIG]
                });
            } else {
                throw switchError;
            }
        }
    }
}

// Get current address
export const getCurrentAddress = async () => {
    try {
        if (!provider || !signer) {
            throw new Error('Web3未初始化');
        }
        return await signer.getAddress();
    } catch (error) {
        console.error('获取地址失败:', error);
        throw error;
    }
};

// 获取用户总成就数量
export async function getUserTotalAchievements(address) {
    if (!provider) {
        throw new Error('Web3未初始化');
    }
    return await provider.getUserTotalAchievements(address);
}

// 获取用户特定类型的成就数量
export async function getUserAchievementsByType(address, achievementType) {
    if (!provider) {
        throw new Error('Web3未初始化');
    }
    return await provider.getUserAchievementsByType(address, achievementType);
}

// 检查用户是否拥有特定成就
export async function hasAchievement(address, achievementType) {
    if (!provider) {
        throw new Error('Web3未初始化');
    }
    return await provider.hasAchievement(address, achievementType);
}

// 获取成就详情
export async function getAchievement(tokenId) {
    if (!provider) {
        throw new Error('Web3未初始化');
    }
    return await provider.getAchievement(tokenId);
}

// Get user achievements
export const getUserAllAchievements = async (address) => {
    try {
        if (!provider) {
            throw new Error('Web3未初始化');
        }
        // 这里添加获取用户NFT成就的逻辑
        return []; // 临时返回空数组
    } catch (error) {
        console.error('获取用户成就失败:', error);
        throw error;
    }
};

// 铸造成就（仅管理员）
export async function mintAchievement(to, name, description, achievementType) {
    if (!provider) {
        throw new Error('Web3未初始化');
    }
    
    const tx = await provider.mintAchievement(to, name, description, achievementType);
    return await tx.wait();
}

// 批量铸造成就（仅管理员）
export async function batchMintAchievement(recipients, name, description, achievementType) {
    if (!provider) {
        throw new Error('Web3未初始化');
    }
    
    const tx = await provider.batchMintAchievement(recipients, name, description, achievementType);
    return await tx.wait();
}

// Subscribe to wallet events
export const subscribeToWalletEvents = ({ onAccountsChanged, onChainChanged, onDisconnect }) => {
    if (typeof window !== 'undefined' && window.ethereum) {
        window.ethereum.on('accountsChanged', onAccountsChanged);
        window.ethereum.on('chainChanged', onChainChanged);
        window.ethereum.on('disconnect', onDisconnect);
    }
};

// Unsubscribe from wallet events
export const unsubscribeFromWalletEvents = () => {
    if (typeof window !== 'undefined' && window.ethereum) {
        window.ethereum.removeAllListeners('accountsChanged');
        window.ethereum.removeAllListeners('chainChanged');
        window.ethereum.removeAllListeners('disconnect');
    }
};

// Subscribe to achievement events
export const subscribeToAchievementEvents = ({ onAchievementMinted }) => {
    // 这里添加订阅成就事件的逻辑
};

// Unsubscribe from achievement events
export const unsubscribeFromAchievementEvents = () => {
    // 这里添加取消订阅成就事件的逻辑
}; 