<template>
	<view class="container">
		<!-- 背景装饰 -->
		<view class="decoration">
			<view class="circle circle-1"></view>
			<view class="circle circle-2"></view>
			<view class="dots dots-1"></view>
			<view class="dots dots-2"></view>
		</view>
		
		<view class="content">
			<!-- Logo区域 -->
			<view class="logo-container">
				<image class="logo" src="/static/logo.png" mode="aspectFit" ></image>
			</view>
			
			<!-- 项目名称 -->
			<view class="title-container">
				<text class="title">paperPlay</text>
				<text class="subtitle">探索学术的乐趣</text>
			</view>

			<!-- 登录/注册表单 -->
			<view class="form-container">
				<view class="form-header">
					<text class="form-type" :class="{ active: isLogin }" @click="isLogin = true">登录</text>
					<text class="form-type" :class="{ active: !isLogin }" @click="isLogin = false">注册</text>
				</view>

				<view class="form-content">
					<!-- 邮箱 -->
					<input 
						class="input"
						type="email"
						v-model="formData.email"
						placeholder="邮箱"
					/>
					
					<!-- 密码 -->
					<input 
						class="input"
						type="password"
						v-model="formData.password"
						placeholder="密码"
					/>
					
					<!-- 注册时显示的额外字段 -->
					<template v-if="!isLogin">
						<input 
							class="input"
							type="password"
							v-model="formData.confirmPassword"
							placeholder="确认密码"
						/>
						<input 
							class="input"
							type="text"
							v-model="formData.display_name"
							placeholder="显示名称"
						/>
					</template>
				</view>

				<button class="submit-btn" @click="handleSubmit">
					{{ isLogin ? '登录' : '注册' }}
				</button>
			</view>
		</view>
	</view>
</template>

<script>
import { login, register } from '@/api/auth'

export default {
	data() {
		return {
			isLogin: true,
			formData: {
				email: 'zsh@123.com',
				password: 'admin123456',
				confirmPassword: '',
				display_name: ''
			}
		}
	},
	methods: {
		async handleSubmit() {
			try {
				if (this.isLogin) {
					// 登录前的表单验证
					if (!this.formData.email || !this.formData.password) {
						uni.showToast({
							title: '请填写完整的登录信息',
							icon: 'none'
						})
						return
					}

					// 登录
					const response = await login({
						email: this.formData.email,
						password: this.formData.password
					})
					
					if (response.access_token) {
						// 保存token和用户信息
						uni.setStorageSync('token', response.access_token)
						uni.setStorageSync('refresh_token', response.refresh_token)
						uni.setStorageSync('userInfo', response.user)
						uni.setStorageSync('token_expires_in', response.expires_in)
						
						uni.showToast({
							title: '登录成功',
							icon: 'success'
						})
						
						// 登录成功后跳转到地图页
						uni.reLaunch({
							url: '/pages/map/map'
						})
					} else {
						uni.showToast({
							title: '登录失败',
							icon: 'none'
						})
					}
				} else {
					// 注册前验证
					if (!this.formData.email || !this.formData.password || !this.formData.display_name) {
						uni.showToast({
							title: '请填写完整的注册信息',
							icon: 'none'
						})
						return
					}

					if (this.formData.password !== this.formData.confirmPassword) {
						uni.showToast({
							title: '两次密码不一致',
							icon: 'none'
						})
						return
					}
					
					// 注册
					const response = await register({
						email: this.formData.email,
						password: this.formData.password,
						display_name: this.formData.display_name
					})
					
					if (response.access_token) {
						// 保存token和用户信息
						uni.setStorageSync('token', response.access_token)
						uni.setStorageSync('refresh_token', response.refresh_token)
						uni.setStorageSync('userInfo', response.user)
						uni.setStorageSync('token_expires_in', response.expires_in)
						
						uni.showToast({
							title: '注册成功',
							icon: 'success'
						})
						
						// 注册成功后直接进入地图页
						uni.reLaunch({
							url: '/pages/map/map'
						})
					} else {
						uni.showToast({
							title: '注册失败',
							icon: 'none'
						})
					}
				}
			} catch (error) {
				console.error('操作失败:', error)
				// 处理错误响应
				uni.showToast({
					title: error.message || '操作失败',
					icon: 'none'
				})
				
				// 如果是未授权错误，清除所有存储的信息
				if (error.message === '请重新登录') {
					uni.removeStorageSync('token')
					uni.removeStorageSync('refresh_token')
					uni.removeStorageSync('userInfo')
					uni.removeStorageSync('token_expires_in')
				}
			}
		}
	}
}
</script>

<style>
.container {
	min-height: 100vh;
	background: linear-gradient(135deg, #fbf0d8 0%, #fff5e5 100%);
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: space-between;
	padding: 80px 0 40px;
	box-sizing: border-box;
	position: relative;
	overflow: hidden;
}

.decoration {
	position: absolute;
	width: 100%;
	height: 100%;
	pointer-events: none;
}

.circle {
	position: absolute;
	border-radius: 50%;
	opacity: 0.1;
}

.circle-1 {
	width: 300px;
	height: 300px;
	background: #ff9966;
	top: -100px;
	right: -100px;
}

.circle-2 {
	width: 200px;
	height: 200px;
	background: #ff7f50;
	bottom: -50px;
	left: -50px;
}

.dots {
	position: absolute;
	width: 100px;
	height: 100px;
	background-image: radial-gradient(#666 2px, transparent 2px);
	background-size: 15px 15px;
	opacity: 0.1;
}

.dots-1 {
	top: 20%;
	left: 10%;
}

.dots-2 {
	bottom: 30%;
	right: 10%;
}

.content {
	display: flex;
	flex-direction: column;
	align-items: center;
	width: 100%;
	z-index: 1;
}

.logo-container {
	position: relative;
	margin-bottom: 40px;
}

.logo {
	width: 120px;
	height: 120px;
	border-radius: 30px;
	box-shadow: 0 8px 20px rgba(0, 0, 0, 0.1);
	background: white;
	/* padding: 15px; */
	animation: float 3s ease-in-out infinite;
}



.title-container {
	text-align: center;
	animation: fadeIn 1s ease-out;
}

.title {
	font-size: 36px;
	font-weight: 600;
	color: #333333;
	margin-bottom: 12px;
	display: block;
	text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.1);
}

.subtitle {
	font-size: 18px;
	color: #666666;
	display: block;
	margin-bottom: 30px;
}

.feature-list {
	display: flex;
	justify-content: center;
	gap: 20px;
	margin-top: 30px;
}

.feature-item {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 8px;
	animation: slideUp 0.5s ease-out backwards;
}

.feature-item:nth-child(2) {
	animation-delay: 0.2s;
}

.feature-item:nth-child(3) {
	animation-delay: 0.4s;
}

.feature-icon {
	font-size: 24px;
}

.feature-text {
	font-size: 14px;
	color: #666666;
}

.button-container {
	width: 100%;
	padding: 0 32px;
	box-sizing: border-box;
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 12px;
}

.start-btn {
	background: linear-gradient(135deg, #ff9966 0%, #ff7f50 100%);
	width: 100%;
	height: 56px;
	border-radius: 28px;
	color: #ffffff;
	font-size: 18px;
	display: flex;
	align-items: center;
	justify-content: center;
	transition: all 0.3s ease;
	padding: 0;
	margin: 0;
	line-height: 1;
	box-shadow: 0 4px 15px rgba(255, 127, 80, 0.3);
}

.start-btn:active {
	transform: scale(0.98);
	box-shadow: 0 2px 8px rgba(255, 127, 80, 0.2);
}

.arrow {
	margin-left: 8px;
	font-size: 20px;
	transition: transform 0.3s ease;
}

.start-btn:active .arrow {
	transform: translateX(4px);
}

.hint-text {
	font-size: 14px;
	color: #999;
	opacity: 0.8;
}

@keyframes float {
	0%, 100% { transform: translateY(0); }
	50% { transform: translateY(-10px); }
}

@keyframes shine {
	0%, 100% { transform: translateX(-100%) rotate(45deg); }
	50% { transform: translateX(100%) rotate(45deg); }
}

@keyframes fadeIn {
	from { opacity: 0; transform: translateY(20px); }
	to { opacity: 1; transform: translateY(0); }
}

@keyframes slideUp {
	from { opacity: 0; transform: translateY(20px); }
	to { opacity: 1; transform: translateY(0); }
}

/* 表单样式 */
.form-container {
	width: 80%;
	max-width: 600rpx;
	margin-top: 60rpx;
}

.form-header {
	display: flex;
	justify-content: center;
	gap: 60rpx;
	margin-bottom: 40rpx;
}

.form-type {
	font-size: 32rpx;
	color: #666;
	padding: 10rpx 20rpx;
	cursor: pointer;
}

.form-type.active {
	color: #ff7f50;
	border-bottom: 4rpx solid #ff7f50;
}

.form-content {
	display: flex;
	flex-direction: column;
	gap: 20rpx;
}

.input {
	height: 80rpx;
	background: rgba(255, 255, 255, 0.8);
	border: 2rpx solid #ddd;
	border-radius: 40rpx;
	padding: 0 30rpx;
	font-size: 28rpx;
}

.submit-btn {
	margin-top: 40rpx;
	height: 88rpx;
	background: linear-gradient(135deg, #ff9966 0%, #ff7f50 100%);
	border-radius: 44rpx;
	color: white;
	font-size: 32rpx;
	border: none;
	box-shadow: 0 4rpx 12rpx rgba(255, 127, 80, 0.3);
}

.submit-btn:active {
	transform: scale(0.98);
}
</style>
