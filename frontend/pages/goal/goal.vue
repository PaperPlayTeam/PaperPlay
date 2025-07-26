<template>
	<view class="container">
		<!-- 顶部凸弧背景 -->
		<view class="header-container">
			<svg class="bottom-arch" viewBox="0 0 750 120" preserveAspectRatio="none">
				<path d="M0,0 C300,60 450,60 750,0 L750,120 L0,120 Z" fill="#b5d6f2" />
				<text x="375" y="50" text-anchor="middle" dominant-baseline="middle" class="app-title">成就</text>
			</svg>
		</view>

		<!-- 内容区域 -->
		<scroll-view class="content-area" scroll-y>
			<view class="content-wrapper">
				<!-- 钱包连接区域 -->
				<view class="wallet-section">
					<view class="wallet-card" v-if="!walletConnected">
						<view class="wallet-info">
							<text class="wallet-title">🔗 区块链成就</text>
							<text class="wallet-description">连接钱包查看您的NFT成就</text>
						</view>
						<button 
							class="connect-btn" 
							@click="connectWallet"
							:loading="isConnectingWallet"
						>
							{{ isConnectingWallet ? '连接中...' : '连接钱包' }}
						</button>
					</view>
					
					<view class="wallet-card connected" v-else>
						<view class="wallet-info">
							<text class="wallet-title">✅ 钱包已连接</text>
							<text class="wallet-address">{{ formatAddress(walletAddress) }}</text>
							<text class="wallet-network">{{ networkName || 'Injective Testnet' }}</text>
						</view>
						<button class="disconnect-btn" @click="disconnectWallet">断开</button>
					</view>
				</view>

				<view class="achievement-stats">
					<view class="stat-card" id="total-questions">
						<text class="stat-number">{{ userStats.totalQuestions }}</text>
						<text class="stat-label">完成题目</text>
					</view>
					<view class="stat-card" id="streak-days">
						<text class="stat-number">{{ userStats.streakDays }}</text>
						<text class="stat-label">连续天数</text>
					</view>
					<view class="stat-card" id="total-points">
						<text class="stat-number">{{ userStats.totalPoints }}</text>
						<text class="stat-label">总积分</text>
					</view>
					<view class="stat-card" id="total-nfts" v-if="walletConnected">
						<text class="stat-number">{{ userStats.totalNFTs }}</text>
						<text class="stat-label">NFT成就</text>
					</view>
				</view>

				<view v-if="loading" class="loading">
					<text>加载中...</text>
				</view>

				<view v-else>
					<!-- 区块链成就区域 -->
					<view v-if="walletConnected && blockchainAchievements.length > 0" class="section-title">
						<text>🏆 链上NFT成就</text>
					</view>
					
					<view v-if="walletConnected && blockchainAchievements.length > 0" class="blockchain-achievements">
						<view 
							v-for="nft in blockchainAchievements" 
							:key="nft.tokenId"
							class="achievement-card blockchain-achievement"
						>
							<view class="achievement-icon">
								<view class="nft-badge">
									<text class="nft-icon">{{ getNFTIcon(nft.achievementType) }}</text>
									<view class="nft-verified">
										<text class="verified-text">✓</text>
									</view>
								</view>
							</view>
							<view class="achievement-info">
								<view class="achievement-header">
									<text class="achievement-title">{{ nft.name }}</text>
									<view class="nft-tag">
										<text>NFT #{{ nft.tokenId }}</text>
									</view>
								</view>
								<text class="achievement-description">{{ nft.description }}</text>
								<view class="achievement-details">
									<text class="achievement-type">类型: {{ getAchievementTypeName(nft.achievementType) }}</text>
									<text class="achievement-time">铸造: {{ formatDate(nft.timestamp) }}</text>
								</view>
							</view>
						</view>
					</view>

					<!-- 传统成就区域 -->
					<view class="section-title">
						<text>📋 系统成就</text>
					</view>
					<view class="achievement-grid">
						<view 
							v-for="(achievement, index) in achievements" 
							:key="achievement.id"
							class="achievement-card"
							:class="{ 'unlocked': achievement.unlocked, 'blockchain-linked': achievement.nftMinted }"
						>
							<view class="achievement-icon">
								<svg :id="`achievement-${index}`" width="80" height="80"></svg>
								<view v-if="achievement.nftMinted" class="blockchain-indicator">
									<text>🔗</text>
								</view>
							</view>
							<view class="achievement-info">
								<view class="achievement-header">
									<text class="achievement-title">{{ achievement.title }}</text>
									<view v-if="achievement.nftMinted" class="nft-linked-tag">
										<text>已上链</text>
									</view>
								</view>
								<text class="achievement-description">{{ achievement.description }}</text>
								<view class="achievement-progress">
									<text>{{ achievement.progress }}/{{ achievement.total }}</text>
									<text v-if="achievement.earnedAt" class="achievement-date">{{ formatDate(achievement.earnedAt) }}</text>
								</view>
							</view>
						</view>
					</view>
				</view>
			</view>
		</scroll-view>

		<!-- 底部docker栏 -->
		<view class="bottom-bg">
			<image src="/static/docker/梯形.svg" mode="aspectFill"></image>
		</view>

		<view class="bottom-docker">
			<view class="trapezoid t1" @click="navigateTo('map')">
				<image style="z-index: 20;margin-left: 40rpx;" src="/static/docker/地图.svg" class="docker-icon" mode="aspectFit"></image>
			</view>
			<view class="trapezoid t2" @click="navigateTo('daily')">
				<image style="z-index: 20;margin-left: 40rpx;" src="/static/docker/日推.svg" class="docker-icon" mode="aspectFit"></image>
			</view>
			<view class="trapezoid t3 active" @click="navigateTo('achievement')">
				<image style="z-index: 20;margin-left: 40rpx;" src="/static/docker/成就.svg" class="docker-icon" mode="aspectFit"></image>
			</view>
			<view class="trapezoid t4" @click="navigateTo('profile')">
				<image style="z-index: 20;margin-left: 40rpx;" src="/static/docker/我.svg" class="docker-icon" mode="aspectFit"></image>
			</view>
		</view>
	</view>
</template>

<script>
import rough from 'roughjs'
import { getAllAchievements, getUserAchievements } from '@/api/achievements'
import { 
    initWeb3, 
    getCurrentAddress, 
    getUserAllAchievements, 
    subscribeToWalletEvents, 
    subscribeToAchievementEvents,
    unsubscribeFromWalletEvents,
    unsubscribeFromAchievementEvents
} from '@/utils/web3.js'
import { ethers } from 'ethers';
export default {
	data() {
		return {
			achievements: [],
			userStats: {
				totalQuestions: 0,
				streakDays: 0,
				totalPoints: 0,
				totalAchievements: 0,
				totalNFTs: 0
			},
			loading: false,
			// 区块链相关状态
			web3Initialized: false,
			walletConnected: false,
			walletAddress: null,
			blockchainAchievements: [],
			isConnectingWallet: false,
			chainId: null,
			networkName: null
		}
	},
	mounted() {
		this.$nextTick(() => {
			this.loadAchievements()
			
			// 添加区块链事件监听
			this.setupBlockchainEvents()
		})
	},
	beforeDestroy() {
		// 清理事件监听
		unsubscribeFromWalletEvents()
		unsubscribeFromAchievementEvents()
	},
	methods: {
		async loadAchievements() {
			try {
				this.loading = true
				
				// 获取所有成就和用户成就
				const [allAchievementsRes, userAchievementsRes] = await Promise.all([
					getAllAchievements(),
					getUserAchievements()
				])
				
				// 获取用户已获得的成就ID列表
				const userAchievementIds = (userAchievementsRes.data || []).map(ua => ua.achievement_id)
				
				// 处理成就数据
				this.achievements = allAchievementsRes.data.map(achievement => {
					// 检查是否是"首战告捷"成就，如果是则强制设置为已解锁
					const isFirstVictory = achievement.id === 'f1ee1a1f-77ea-4df8-b8e4-d214512140bf'
					const isUnlocked = isFirstVictory ? true : userAchievementIds.includes(achievement.id)
					
					const userAchievement = userAchievementsRes.data?.find(ua => ua.achievement_id === achievement.id)
					
					// 根据成就等级设置不同的样式
					const levelStyles = {
						1: { color: '#CD7F32', name: 'Bronze' }, // 铜牌
						2: { color: '#C0C0C0', name: 'Silver' }, // 银牌
						3: { color: '#FFD700', name: 'Gold' }    // 金牌
					}
					
					// 获取成就获得时间
					const earnedAt = isFirstVictory ? new Date() : (userAchievement ? new Date(userAchievement.earned_at) : null)
					
					return {
						id: achievement.id,
						title: achievement.name,
						description: achievement.description,
						iconUrl: achievement.icon_url,
						level: achievement.level,
						unlocked: isUnlocked,
						earnedAt: earnedAt,
						levelStyle: levelStyles[achievement.level] || { color: '#ff7f50', name: 'Default' },
						nftMetadata: achievement.nft_metadata,
						progress: isUnlocked ? 1 : 0,
						total: 1
					}
				})
				
				// 更新用户统计数据，确保包含"首战告捷"
				this.userStats = {
					...this.userStats,
					totalAchievements: Math.max(1, userAchievementIds.length) // 至少有一个成就
				}
				
				// 如果钱包已连接，加载区块链成就
				if (this.walletConnected) {
					await this.loadBlockchainAchievements()
				}
				
				// 绘制成就图标
				this.$nextTick(() => {
					this.drawAchievementIcons()
				})
			} catch (error) {
				console.error('加载成就失败:', error)
				uni.showToast({
					title: error.message || '加载成就失败',
					icon: 'none'
				})
			} finally {
				this.loading = false
			}
		},
		
		// 钱包连接方法
		async connectWallet() {
			if (this.isConnectingWallet) return;
			
			this.isConnectingWallet = true;
			try {
				// Initialize Web3
				const { provider, signer } = await initWeb3();
				this.web3Initialized = true;
				
				// Get address
				const address = await signer.getAddress();
				this.walletAddress = address;
				this.walletConnected = true;
				
				// Get network info
				const network = await provider.getNetwork();
				this.chainId = network.chainId.toString();
				this.networkName = 'Injective Testnet';
				
				// Load achievements
				await this.loadBlockchainAchievements();
				
				uni.showToast({
					title: '钱包连接成功',
					icon: 'success'
				});
			} catch (error) {
				console.error('连接钱包失败:', error);
				uni.showToast({
					title: error.message || '连接钱包失败',
					icon: 'none'
				});
				this.web3Initialized = false;
				this.walletConnected = false;
			} finally {
				this.isConnectingWallet = false;
			}
		},
		
		// 断开钱包连接
		disconnectWallet() {
			this.walletConnected = false
			this.walletAddress = null
			this.blockchainAchievements = []
			this.userStats.totalNFTs = 0
			this.web3Initialized = false
			
			uni.showToast({
				title: '钱包已断开',
				icon: 'success'
			})
		},
		
		// 加载区块链成就
		async loadBlockchainAchievements() {
			if (!this.walletAddress) return
			
			try {
				const blockchainAchievements = await getUserAllAchievements(this.walletAddress)
				this.blockchainAchievements = blockchainAchievements
				this.userStats.totalNFTs = blockchainAchievements.length
				
				// 标记已上链的成就
				this.achievements = this.achievements.map(achievement => {
					const isMinted = blockchainAchievements.some(nft => 
						nft.name === achievement.title || 
						nft.description === achievement.description
					)
					return {
						...achievement,
						nftMinted: isMinted
					}
				})
				
			} catch (error) {
				console.error('加载区块链成就失败:', error)
				// 静默处理错误，不影响传统成就显示
			}
		},
		
		// 设置区块链事件监听
		setupBlockchainEvents() {
			subscribeToWalletEvents({
				onAccountsChanged: (accounts) => {
					if (accounts.length === 0) {
						this.disconnectWallet()
					} else {
						this.walletAddress = accounts[0]
						this.loadBlockchainAchievements()
					}
				},
				onChainChanged: (chainId) => {
					this.chainId = chainId
					if (this.walletConnected) {
						this.loadBlockchainAchievements()
					}
				},
				onDisconnect: () => {
					this.disconnectWallet()
				}
			})
			
			subscribeToAchievementEvents({
				onAchievementMinted: () => {
					// 重新加载区块链成就
					this.loadBlockchainAchievements()
					uni.showToast({
						title: '新NFT成就已铸造',
						icon: 'success'
					})
				}
			})
		},
		
		// 格式化地址
		formatAddress(address) {
			if (!address) return ''
			return address.substring(0, 6) + '...' + address.substring(address.length - 4)
		},
		
		// 格式化日期
		formatDate(date) {
			if (!date) return ''
			const d = new Date(date)
			return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
		},
		
		// 获取成就类型名称
		getAchievementTypeName(type) {
			const types = {
				0: '青铜',
				1: '白银', 
				2: '黄金',
				3: '钻石'
			}
			return types[type] || '未知'
		},
		
		// 原有的绘制成就图标方法保持不变
		drawAchievementIcons() {
			this.achievements.forEach((achievement, index) => {
				const svg = document.getElementById(`achievement-${index}`)
				if (svg) {
					// 清除现有内容
					while (svg.firstChild) {
						svg.removeChild(svg.firstChild)
					}
					
					const rc = rough.svg(svg)
					
					// 绘制成就徽章外框
					const badge = rc.circle(40, 40, 70, {
						stroke: achievement.unlocked ? achievement.levelStyle.color : '#ccc',
						strokeWidth: 2,
						fill: achievement.unlocked ? '#fff1e6' : '#f5f5f5',
						roughness: 1.5,
						bowing: 1
					})
					svg.appendChild(badge)
					
					// 为已解锁的成就绘制星星
					if (achievement.unlocked) {
						// 绘制发光效果
						const glow = rc.circle(40, 40, 75, {
							stroke: achievement.levelStyle.color,
							strokeWidth: 1,
							fill: 'none',
							roughness: 0.5,
							bowing: 1
						})
						svg.appendChild(glow)
						
						// 绘制星星
						const star = rc.path('M40,15 L45,30 L60,30 L48,40 L52,55 L40,47 L28,55 L32,40 L20,30 L35,30 Z', {
							fill: achievement.levelStyle.color,
							stroke: achievement.levelStyle.color,
							strokeWidth: 1,
							roughness: 1.2,
							fillStyle: 'solid'
						})
						svg.appendChild(star)
						
						// 添加小装饰点
						const dots = [
							[60, 20], [20, 20], [40, 10],
							[65, 40], [15, 40], [40, 70]
						]
						dots.forEach(([x, y]) => {
							const dot = rc.circle(x, y, 3, {
								fill: achievement.levelStyle.color,
								stroke: 'none',
								roughness: 0.8,
								fillStyle: 'solid'
							})
							svg.appendChild(dot)
						})
					}
				}
			})
		},
		navigateTo(page) {
			const routes = {
				map: '/pages/map/map',
				daily: '/pages/daily/daily',
				achievement: '/pages/goal/goal',
				profile: '/pages/profile/profile'
			}
			
			if (page !== 'achievement') {
				uni.navigateTo({
					url: routes[page]
				})
			}
		}
	}
}
</script>

<style>
@font-face {
    font-family: 'PingFang';
    src: url('/static/fonts/平方手书体.ttf') format('truetype');
    font-weight: normal;
    font-style: normal;
    font-display: swap;
}

.container {
	width: 100vw;
	height: 100vh;
	background-color: #b5d6f2;
	position: relative;
	overflow: hidden;
}

/* 顶部凸弧背景 */
.header-container {
	width: 100%;
	height: 120rpx;
	background-color: #b5d6f2;
	position: relative;
}

.bottom-arch {
	width: 100%;
	height: 100%;
	display: block;
	position: absolute;
	top: 0;
	left: 0;
}

.app-title {
	font-family: 'PingFang', cursive;
	font-size: 60rpx;
	color: #ffffff;
	font-weight: bold;
	text-shadow: 2rpx 2rpx 4rpx rgba(0, 0, 0, 0.2);
}

/* 内容区域 */
.content-area {
	position: absolute;
	top: 120rpx;
	left: 0;
	right: 0;
	bottom: 160rpx;
	padding: 40rpx;
	box-sizing: border-box;
}

.content-wrapper {
	display: flex;
	flex-direction: column;
	gap: 40rpx; /* Added gap for spacing between sections */
}

.achievement-stats {
	display: flex;
	justify-content: space-between;
	margin-bottom: 60rpx;
	gap: 20rpx;
}

.stat-card {
	background: white;
	border-radius: 24rpx;
	padding: 30rpx;
	flex: 1;
	text-align: center;
	box-shadow: 0 4rpx 16rpx rgba(0, 0, 0, 0.1);
}

.stat-number {
	display: block;
	font-family: 'PingFang';
	font-size: 48rpx;
	font-weight: bold;
	color: #ff7f50;
	margin-bottom: 10rpx;
}

.stat-label {
	display: block;
	font-size: 28rpx;
	color: #666;
}

.achievement-grid {
	display: flex;
	flex-direction: column;
	gap: 30rpx;
}

.achievement-card {
	display: flex;
	align-items: center;
	background: white;
	border-radius: 24rpx;
	padding: 30rpx;
	gap: 30rpx;
	box-shadow: 0 4rpx 16rpx rgba(0, 0, 0, 0.1);
	transition: transform 0.3s ease;
}

.achievement-card:active {
	transform: scale(0.98);
}

.achievement-card.unlocked {
	border-left: 8rpx solid #ff7f50;
}

.achievement-icon {
	flex-shrink: 0;
}

.achievement-info {
	flex: 1;
}

.achievement-title {
	display: block;
	font-family: 'PingFang';
	font-size: 36rpx;
	font-weight: bold;
	color: #3e2a1c;
	margin-bottom: 10rpx;
}

.achievement-description {
	display: block;
	font-size: 28rpx;
	color: #666;
	margin-bottom: 15rpx;
	line-height: 1.4;
}

.achievement-progress {
	display: flex;
	align-items: center;
	gap: 10rpx;
}

.achievement-progress text {
	font-size: 24rpx;
	color: #ff7f50;
	font-weight: bold;
}

/* 底部背景 */
.bottom-bg {
	position: fixed;
	bottom: 0;
	left: 0;
	right: 0;
	height: 240rpx;
	z-index: 9995;
}

.bottom-bg image {
	width: 100%;
	height: 100%;
}

/* 底部docker栏 */
.bottom-docker {
	position: fixed;
	bottom: 0;
	left: 0;
	right: 0;
	height: 160rpx;
	display: flex;
	z-index: 9996;
}

.trapezoid {
	flex: 1;
	height: 160rpx;
	position: relative;
	margin-left: -40rpx;
}

.trapezoid:first-child {
	margin-left: 0;
}

.trapezoid.active .docker-icon {
	opacity: 1;
}

.trapezoid:not(.active) .docker-icon {
	opacity: 0.6;
}

.docker-icon {
	position: absolute;
	width: 84rpx;
	height: 84rpx;
	left: 50%;
	bottom: 30rpx;
	transform: translateX(-50%);
	z-index: 2;
}

.loading {
	display: flex;
	justify-content: center;
	align-items: center;
	padding: 40rpx;
}

.loading text {
	font-size: 28rpx;
	color: #666;
}

/* 区块链相关样式 */
.wallet-section {
	margin-bottom: 40rpx;
}

.wallet-card {
	display: flex;
	align-items: center;
	justify-content: space-between;
	background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
	border-radius: 24rpx;
	padding: 40rpx;
	box-shadow: 0 8rpx 24rpx rgba(0, 0, 0, 0.15);
	transition: transform 0.3s ease;
}

.wallet-card.connected {
	background: linear-gradient(135deg, #56ab2f 0%, #a8e6cf 100%);
}

.wallet-info {
	flex: 1;
}

.wallet-title {
	display: block;
	font-family: 'PingFang';
	font-size: 36rpx;
	font-weight: bold;
	color: #ffffff;
	margin-bottom: 10rpx;
}

.wallet-description, .wallet-address, .wallet-network {
	display: block;
	font-size: 28rpx;
	color: rgba(255, 255, 255, 0.9);
	margin-bottom: 5rpx;
}

.connect-btn, .disconnect-btn {
	background: rgba(255, 255, 255, 0.2);
	color: #ffffff;
	border: 2rpx solid rgba(255, 255, 255, 0.3);
	border-radius: 24rpx;
	padding: 20rpx 40rpx;
	font-size: 28rpx;
	transition: all 0.3s ease;
}

.connect-btn:hover, .disconnect-btn:hover {
	background: rgba(255, 255, 255, 0.3);
	transform: translateY(-2rpx);
}

.section-title {
	margin: 40rpx 0 30rpx;
	padding: 0 20rpx;
}

.section-title text {
	font-family: 'PingFang';
	font-size: 40rpx;
	font-weight: bold;
	color: #3e2a1c;
}

.blockchain-achievements {
	margin-bottom: 40rpx;
}

.blockchain-achievement {
	background: linear-gradient(135deg, #ffecd2 0%, #fcb69f 100%);
	border-left: 8rpx solid #ff7f50;
}

.nft-badge {
	position: relative;
	display: flex;
	align-items: center;
	justify-content: center;
	width: 80rpx;
	height: 80rpx;
	background: linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%);
	border-radius: 50%;
	box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.15);
}

.nft-icon {
	font-size: 40rpx;
}

.nft-verified {
	position: absolute;
	top: -10rpx;
	right: -10rpx;
	background: #4caf50;
	border-radius: 50%;
	width: 30rpx;
	height: 30rpx;
	display: flex;
	align-items: center;
	justify-content: center;
}

.verified-text {
	color: white;
	font-size: 20rpx;
	font-weight: bold;
}

.nft-tag, .nft-linked-tag {
	background: #ff7f50;
	color: white;
	padding: 8rpx 16rpx;
	border-radius: 16rpx;
	font-size: 24rpx;
	font-weight: bold;
}

.achievement-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	margin-bottom: 10rpx;
}

.achievement-details {
	display: flex;
	flex-direction: column;
	gap: 5rpx;
	margin-top: 10rpx;
}

.achievement-type, .achievement-time {
	font-size: 24rpx;
	color: #666;
}

.achievement-date {
	font-size: 24rpx;
	color: #999;
	margin-left: 20rpx;
}

.blockchain-indicator {
	position: absolute;
	top: -10rpx;
	right: -10rpx;
	background: #4285f4;
	border-radius: 50%;
	width: 40rpx;
	height: 40rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 20rpx;
	box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.2);
}

.achievement-card.blockchain-linked {
	border-left: 8rpx solid #4285f4;
	background: linear-gradient(135deg, #e3f2fd 0%, #bbdefb 100%);
}

@media (prefers-color-scheme: dark) {
	.wallet-card {
		background: linear-gradient(135deg, #434343 0%, #000000 100%);
	}
	
	.wallet-card.connected {
		background: linear-gradient(135deg, #2e7d32 0%, #66bb6a 100%);
	}
	
	.blockchain-achievement {
		background: linear-gradient(135deg, #3e2723 0%, #5d4037 100%);
	}
}
</style>