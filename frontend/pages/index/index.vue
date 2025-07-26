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
				<image class="logo" src="/static/login/Group.svg" mode="aspectFit" ></image>
			</view>
			


			<!-- 登录/注册表单 -->
			<view class="form-section">
				<view class="form-container">
			
					<view class="form-content">
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
			<view class="bottom-section">
				<view class="bottom-images">
					<image class="bottom-img" src="/static/login/Union.png" mode="aspectFit"></image>
					<image class="bottom-img" src="/static/login/Union-1.png" mode="aspectFit"></image>
					<image class="bottom-img" src="/static/login/Union-2.png" mode="aspectFit"></image>
					<image class="bottom-img" src="/static/login/Union-3.png" mode="aspectFit"></image>
				</view>
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
	width: 100vw;
	background: #fff;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: flex-start; /* 改为 flex-start，让内容从顶部开始 */
	position: relative;
	padding: 0;
	box-sizing: border-box;
}

.logo-container {
	width: 100vw;
	display: flex;
	justify-content: center;
	align-items: center;
	margin-top: 60rpx; /* 从 100rpx 减少到 60rpx */
	margin-bottom: 20rpx;
}

.logo {
	width: 600rpx;  /* 超级大！从 260rpx 增加到 600rpx */
	height: 600rpx; /* 超级大！从 260rpx 增加到 600rpx */
	display: block;
}

.form-section {
	width: 100vw;
	display: flex;
	justify-content: center;
	align-items: center;
	flex: 1;
	min-height: 400rpx;
	margin-top: 40rpx; /* 添加顶部边距 */
}

.form-container {
	width: 90vw;
	max-width: 600rpx;
	background: #fff;
	border: 3px solid #b5d6f2;
	border-radius: 32rpx;
	padding: 60rpx 40rpx 40rpx 40rpx;
	box-sizing: border-box;
	box-shadow: 0 8rpx 32rpx rgba(181, 214, 242, 0.08);
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
}

.form-header {
	width: 100%;
	text-align: center;
	margin-bottom: 40rpx;
}

.form-title {
	font-size: 40rpx;
	color: #3e2a1c;
	font-weight: bold;
	letter-spacing: 2rpx;
}

.form-content {
	width: 100%;
	display: flex;
	flex-direction: column;
	gap: 32rpx;
	margin-bottom: 40rpx;
}

.input {
	width: 100%;
	box-sizing: border-box;
	height: 80rpx;
	background: rgba(245, 245, 245, 0.8);
	border: 2rpx solid #b5d6f2;
	border-radius: 40rpx;
	padding: 0 30rpx;
	font-size: 28rpx;
	color: #333;
}

.submit-btn {
	width: 100%;
	height: 88rpx;
	background: linear-gradient(135deg, #b5d6f2 0%, #7ec0ee 100%);
	border-radius: 44rpx;
	color: #3e2a1c;
	font-size: 32rpx;
	border: none;
	box-shadow: 0 4rpx 12rpx rgba(181, 214, 242, 0.15);
	font-weight: bold;
	letter-spacing: 2rpx;
}

.submit-btn:active {
	transform: scale(0.98);
}

.bottom-section {
	width: 100vw;
	display: flex;
	justify-content: center;
	align-items: center;
	margin-bottom: 60rpx;
}

.bottom-images {
	display: flex;
	justify-content: center;
	align-items: center;
	gap: 32rpx;
	width: 100%;
}

.bottom-img {
	width: 140rpx;
	height: 140rpx;
	object-fit: contain;
}

@media (max-width: 500px) {
	.form-container {
		padding: 40rpx 10rpx 30rpx 10rpx;
	}
	.logo-container {
		margin-top: 30rpx; /* 移动端减少到 30rpx */
		margin-bottom: 10rpx;
	}
	.logo {
		width: 400rpx;
		height: 400rpx;
	}
	.form-section {
		margin-top: 20rpx; /* 移动端减少顶部边距 */
	}
	.bottom-section {
		margin-bottom: 30rpx;
	}
	.bottom-img {
		width: 100rpx;
		height: 100rpx;
	}
}
</style>
