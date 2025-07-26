<template>
	<view class="container">
		<!-- 顶部凸弧背景 -->
		<view class="header-container">
			<svg class="bottom-arch" viewBox="0 0 750 120" preserveAspectRatio="none">
				<path d="M0,0 C300,60 450,60 750,0 L750,120 L0,120 Z" fill="#b5d6f2"/>
				<text x="375" y="50" text-anchor="middle" dominant-baseline="middle" class="app-title">个人中心</text>
			</svg>
		</view>

		<!-- 内容区域 -->
		<scroll-view class="content-area" scroll-y>
			<!-- 用户信息卡片 -->
			<view class="user-profile-card" id="user-card">
				<view class="user-avatar">
					<svg id="avatar-svg" width="120" height="120"></svg>
				</view>
				<view class="user-info">
					<text class="username">学术探险家</text>
					<text class="user-level">Level 5 · 资深学者</text>
					<view class="level-progress">
						<view class="progress-bar" id="progress-bar">
							<view class="progress-fill" :style="{width: userProgress + '%'}"></view>
						</view>
						<text class="progress-text">{{ userLevel }}/{{ totalLevel }}</text>
					</view>
				</view>
			</view>

			<!-- 统计信息 -->
			<view class="stats-section">
				<text class="section-title">学习统计</text>
				<view class="stats-grid">
					<view class="stat-item" id="stat-ai">
						<view class="stat-icon">
							<svg id="ai-icon" width="60" height="60"></svg>
						</view>
						<view class="stat-content">
							<text class="stat-value">15</text>
							<text class="stat-name">AI题目</text>
						</view>
					</view>

				</view>
			</view>
			



			<!-- 关于信息 -->
			<view class="about-section">
				<text class="section-title">关于 paperPlay</text>
				<view class="about-content">
					<text class="about-text">版本：v0.0.1</text>
				</view>
			</view>

			<!-- 退出登录按钮 -->
			<view class="logout-section">
				<button class="logout-button" @click="handleLogout">退出登录</button>
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
			<view class="trapezoid t3" @click="navigateTo('achievement')">
				<image style="z-index: 20;margin-left: 40rpx;" src="/static/docker/成就.svg" class="docker-icon" mode="aspectFit"></image>
			</view>
			<view class="trapezoid t4 active" @click="navigateTo('profile')">
				<image style="z-index: 20;margin-left: 40rpx;" src="/static/docker/我.svg" class="docker-icon" mode="aspectFit"></image>
			</view>
		</view>
	</view>
</template>

<script>
import rough from 'roughjs'

export default {
	data() {
		return {
			userLevel: 5,
			totalLevel: 10,
			userProgress: 75
		}
	},
	mounted() {
		this.$nextTick(() => {
			this.drawUserProfile()
		})
	},
	methods: {
		drawUserProfile() {
			// Draw user avatar
			const avatarSvg = document.getElementById('avatar-svg')
			if (avatarSvg) {
				const rc = rough.svg(avatarSvg)
				
				// Draw avatar circle
				const avatar = rc.circle(60, 60, 100, {
					stroke: '#ff7f50',
					strokeWidth: 3,
					fill: '#fff1e6',
					roughness: 1.5
				})
				avatarSvg.appendChild(avatar)
				
				// Draw user icon
				const userIcon = rc.path('M40,45 Q50,40 60,45 Q70,50 75,65 Q65,75 50,75 Q35,75 25,65 Q30,50 40,45', {
					fill: '#3e2a1c',
					stroke: '#3e2a1c',
					strokeWidth: 1,
					roughness: 1
				})
				avatarSvg.appendChild(userIcon)
			}
			
			// Draw progress bar
			const progressBar = document.getElementById('progress-bar')
			if (progressBar) {
				const rc = rough.svg(progressBar)
				
				const bar = rc.rectangle(0, 0, 200, 10, {
					stroke: '#ccc',
					strokeWidth: 1,
					fill: '#f5f5f5',
					roughness: 1
				})
				progressBar.appendChild(bar)
			}
			
			// Draw category icons
			this.drawCategoryIcon('ai-icon', 'AI', '#ff7f50')
			this.drawCategoryIcon('math-icon', '∑', '#3e2a1c')
			this.drawCategoryIcon('bio-icon', 'DNA', '#a2c69b')
			
			// Draw setting icons
			this.drawSettingIcon('font-icon', 'Aa')
			this.drawSettingIcon('theme-icon', '🎨')
			this.drawSettingIcon('notification-icon', '🔔')
		},
		
		drawCategoryIcon(id, text, color) {
			const svg = document.getElementById(id)
			if (svg) {
				const rc = rough.svg(svg)
				
				const circle = rc.circle(30, 30, 50, {
					stroke: color,
					strokeWidth: 2,
					fill: 'white',
					roughness: 1.5
				})
				svg.appendChild(circle)
				
				const label = rc.text(30, 35, text, {
					fontSize: 20,
					fontFamily: 'PingFang',
					textAlign: 'center',
					fill: color
				})
				svg.appendChild(label)
			}
		},
		
		drawSettingIcon(id, icon) {
			const svg = document.getElementById(id)
			if (svg) {
				const rc = rough.svg(svg)
				
				const rect = rc.rectangle(5, 5, 30, 30, {
					stroke: '#978B6B',
					strokeWidth: 1,
					fill: '#f8f8f8',
					roughness: 1.2
				})
				svg.appendChild(rect)
				
				const label = rc.text(20, 22, icon, {
					fontSize: 16,
					textAlign: 'center',
					fill: '#978B6B'
				})
				svg.appendChild(label)
			}
		},
		
		showToast(message) {
			uni.showToast({
				title: message,
				icon: 'none'
			})
		},
		
		navigateTo(page) {
			const routes = {
				map: '/pages/map/map',
				daily: '/pages/daily/daily',
				achievement: '/pages/goal/goal',
				profile: '/pages/profile/profile'
			}
			
			if (page !== 'profile') {
				uni.navigateTo({
					url: routes[page]
				})
			}
		},

		// 处理退出登录
		handleLogout() {
			uni.showModal({
				title: '确认退出',
				content: '确定要退出登录吗？',
				success: (res) => {
					if (res.confirm) {
						// 清除所有存储的token和用户信息
						uni.removeStorageSync('access_token');
						uni.removeStorageSync('refresh_token');
						uni.removeStorageSync('token_expires_in');
						uni.removeStorageSync('userInfo');
						
						// 显示提示
						uni.showToast({
							title: '已退出登录',
							icon: 'success',
							duration: 2000,
							complete: () => {
								// 延迟跳转，让用户看到提示
								setTimeout(() => {
									// 重定向到登录页
									uni.reLaunch({
										url: '/pages/index/index'
									});
								}, 1000);
							}
						});
					}
				}
			});
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

.user-profile-card {
	display: flex;
	align-items: center;
	background: white;
	border-radius: 32rpx;
	padding: 40rpx;
	margin-bottom: 40rpx;
	box-shadow: 0 8rpx 32rpx rgba(0, 0, 0, 0.1);
	gap: 40rpx;
}

.user-avatar {
	flex-shrink: 0;
}

.user-info {
	flex: 1;
}

.username {
	display: block;
	font-family: 'PingFang';
	font-size: 48rpx;
	font-weight: bold;
	color: #3e2a1c;
	margin-bottom: 15rpx;
}

.user-level {
	display: block;
	font-size: 32rpx;
	color: #666;
	margin-bottom: 20rpx;
}

.level-progress {
	display: flex;
	align-items: center;
	gap: 20rpx;
}

.progress-bar {
	flex: 1;
	height: 10rpx;
	background: #f5f5f5;
	border-radius: 5rpx;
	position: relative;
	overflow: hidden;
}

.progress-fill {
	height: 100%;
	background: linear-gradient(90deg, #ff7f50, #ff9f7f);
	border-radius: 5rpx;
	transition: width 0.3s ease;
}

.progress-text {
	font-size: 24rpx;
	color: #666;
	font-weight: bold;
}

.stats-section,
.settings-section,
.about-section {
	margin-bottom: 40rpx;
}

.section-title {
	font-family: 'PingFang';
	font-size: 44rpx;
	font-weight: bold;
	color: #3e2a1c;
	margin-bottom: 30rpx;
	padding-left: 10rpx;
}

.stats-grid {
	display: flex;
	justify-content: space-between;
	gap: 20rpx;
	margin-bottom: 40rpx;
}

.stat-item {
	background: white;
	border-radius: 24rpx;
	padding: 30rpx;
	flex: 1;
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 15rpx;
	box-shadow: 0 4rpx 16rpx rgba(0, 0, 0, 0.1);
}

.stat-icon {
	margin-bottom: 10rpx;
}

.stat-value {
	font-family: 'PingFang';
	font-size: 36rpx;
	font-weight: bold;
	color: #ff7f50;
}

.stat-name {
	font-size: 28rpx;
	color: #666;
	text-align: center;
}

.settings-list {
	display: flex;
	flex-direction: column;
	gap: 20rpx;
}

.setting-item {
	display: flex;
	align-items: center;
	background: white;
	border-radius: 24rpx;
	padding: 30rpx;
	gap: 20rpx;
	box-shadow: 0 4rpx 16rpx rgba(0, 0, 0, 0.1);
	transition: transform 0.3s ease;
}

.setting-item:active {
	transform: scale(0.98);
}

.setting-icon {
	width: 60rpx;
	height: 60rpx;
}

.setting-name {
	flex: 1;
	font-size: 32rpx;
	color: #3e2a1c;
}

.setting-value {
	font-size: 28rpx;
	color: #666;
}

.about-content {
	background: white;
	border-radius: 24rpx;
	padding: 40rpx;
	box-shadow: 0 4rpx 16rpx rgba(0, 0, 0, 0.1);
}

.about-text {
	display: block;
	font-size: 32rpx;
	color: #666;
	margin-bottom: 20rpx;
	line-height: 1.6;
}

.about-text:last-child {
	margin-bottom: 0;
}

/* 退出登录按钮样式 */
.logout-section {
	margin-top: 40rpx;
	padding: 0 40rpx;
}

.logout-button {
	width: 100%;
	height: 88rpx;
	line-height: 88rpx;
	background: #ff4d4f;
	color: #fff;
	font-size: 32rpx;
	border-radius: 44rpx;
	border: none;
	box-shadow: 0 4rpx 16rpx rgba(255, 77, 79, 0.2);
	transition: all 0.3s ease;
}

.logout-button:active {
	transform: scale(0.98);
	background: #ff7875;
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
</style>