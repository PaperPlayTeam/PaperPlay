<template>
	<view class="container">
		<!-- 顶部凸弧背景 -->
		<view class="header-container">
			<svg class="bottom-arch" viewBox="0 0 750 120" preserveAspectRatio="none">
				<path d="M0,0 C300,60 450,60 750,0 L750,120 L0,120 Z" fill="#b5d6f2"/>
				<text x="375" y="50" text-anchor="middle" dominant-baseline="middle" class="app-title">paperPlay</text>
			</svg>
		</view>

		<!-- 内容区域 -->
		<scroll-view class="content-area" scroll-y>
			<view class="daily-intro">
				<text class="intro-title">每日精选</text>
			</view>

			<!-- 加载状态 -->
			<view v-if="loading" class="loading-container">
				<!-- 骨骼屏动画 -->
				<view class="skeleton-cards">
					<view 
						v-for="i in 3" 
						:key="i"
						class="skeleton-card"
					>
						<view class="skeleton-header">
							<view class="skeleton-number"></view>
							<view class="skeleton-meta">
								<view class="skeleton-tag"></view>
								<view class="skeleton-tag"></view>
							</view>
						</view>
						
						<view class="skeleton-paper-info">
							<view class="skeleton-paper-title"></view>
							<view class="skeleton-paper-author"></view>
						</view>
						
						<view class="skeleton-title"></view>
						<view class="skeleton-title short"></view>
						
						<view class="skeleton-meta">
							<view class="skeleton-score"></view>
						</view>
						
						<view class="skeleton-preview">
							<view class="skeleton-preview-title"></view>
							<view class="skeleton-options">
								<view class="skeleton-option"></view>
								<view class="skeleton-option"></view>
								<view class="skeleton-option"></view>
								<view class="skeleton-option"></view>
							</view>
						</view>
					</view>
				</view>
			</view>
			
			<!-- 问题列表 -->
			<view v-else class="question-cards">
				<view 
					v-for="(question, index) in questions" 
					:key="question.id"
					class="question-card"
					:data-difficulty="question.difficulty"
					@click="startQuiz(question)"
				>
					<view class="card-header">
						<text class="question-number">问题 {{ index + 1 }}</text>
						<view class="question-meta">
							
						</view>
					</view>
					
					<!-- 论文信息 -->
					<view class="paper-info">
						<text class="paper-title">{{ question.paper.title }}</text>
						<text class="paper-author">{{ getCitation(question.paper.citation) }}</text>
					</view>
					
					<view class="card-title">
						<text>{{ question.stem }}</text>
					</view>
					
					<view class="card-meta">
						<text class="score">分值: {{ question.score }}分</text>
					</view>
					
					<view v-if="question.content.type === 'mcq'" class="card-preview">
						<text class="preview-title">选项预览:</text>
						<view class="preview-options">
							<text v-for="(option, optIndex) in question.content.options" 
								:key="optIndex" 
								class="preview-option"
							>
								{{ String.fromCharCode(65 + optIndex) }}
							</text>
						</view>
					</view>
				</view>
			</view>
		</scroll-view>

		<!-- 底部docker栏 -->
		<view class="bottom-bg">
			<image src="/static/docker/梯形.svg" mode="aspectFill"></image>
		</view>

		<!-- 底部docker栏 -->
		<view class="bottom-docker">
			<view class="trapezoid t1" @click="navigateTo('map')">
				<image style="z-index: 20;margin-left: 40rpx;" src="/static/docker/地图.svg" class="docker-icon" mode="aspectFit"></image>
			</view>
			<view class="trapezoid t2 active" @click="navigateTo('daily')">
				<image style="z-index: 20;margin-left: 40rpx;" src="/static/docker/日推.svg" class="docker-icon" mode="aspectFit"></image>
			</view>
			<view class="trapezoid t3" @click="navigateTo('achievement')">
				<image style="z-index: 20;margin-left: 40rpx;" src="/static/docker/成就.svg" class="docker-icon" mode="aspectFit"></image>
			</view>
			<view class="trapezoid t4" @click="navigateTo('profile')">
				<image style="z-index: 20;margin-left: 40rpx;" src="/static/docker/我.svg" class="docker-icon" mode="aspectFit"></image>
			</view>
		</view>
	</view>
</template>

<script>
import { getAllSubjects, getSubjectPapers } from '@/api/subjects';
import { getPaperLevel } from '@/api/papers';
import { getLevelQuestions } from '@/api/levels';

export default {
	data() {
		return {
			papers: [],
			loading: false,
			questions: [],
			aiSubjectId: 'ml_ai_subject'
		}
	},
	async onLoad() {
		// 直接加载AI相关的论文，不需要加载所有学科
		await this.loadAIPapers();
	},
	methods: {
		// 处理API错误
		handleApiError(error) {
			console.error('API错误:', error);
			if (error.message === '请重新登录') {
				// 清除本地存储的token
				uni.removeStorageSync('access_token');
				uni.removeStorageSync('refresh_token');
				uni.removeStorageSync('token_expires_in');
				uni.removeStorageSync('userInfo');
				
				// 显示提示
				uni.showToast({
					title: '请重新登录',
					icon: 'none',
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
			} else {
				// 其他错误显示错误信息
				uni.showToast({
					title: error.message || '发生错误',
					icon: 'none'
				});
			}
		},
		
		// 修改加载方法，直接加载问题列表
		async loadAIPapers() {
			try {
				this.loading = true;
				
				// 获取AI学科的论文
				const response = await getSubjectPapers(this.aiSubjectId);
				if (response.success) {
					this.papers = response.data;
					
					// 从每篇论文中获取问题
					const allQuestions = [];
					for (const paper of this.papers) {
						try {
							// 获取论文的关卡信息
							const levelResponse = await getPaperLevel(paper.id);
							if (levelResponse.success && levelResponse.data) {
								// 获取关卡问题
								const questionsResponse = await getLevelQuestions(levelResponse.data.id);
								if (questionsResponse.success) {
									// 为每个问题添加论文信息和paper_id
									const paperQuestions = questionsResponse.data.map(question => ({
										...question,
										content: JSON.parse(question.content_json || '{}'),
										answer: JSON.parse(question.answer_json || '{}'),
										level_id: levelResponse.data.id,
										paper_id: paper.id, // 添加 paper_id
										paper: {
											id: paper.id,
											title: paper.title,
											citation: paper.citation,
											level: levelResponse.data
										}
									}));
									
									// 随机选择1-2个问题
									const randomQuestions = this.shuffleArray(paperQuestions).slice(0, Math.floor(Math.random() * 2) + 1);
									allQuestions.push(...randomQuestions);
								}
							}
						} catch (error) {
							console.error(`获取论文${paper.id}的问题失败:`, error);
						}
					}
					
					// 随机打乱所有问题，并只取前10个
					this.questions = this.shuffleArray(allQuestions).slice(0, 10);
				}
			} catch (error) {
				this.handleApiError(error);
			} finally {
				this.loading = false;
			}
		},
		
		// 添加数组随机打乱方法
		shuffleArray(array) {
			for (let i = array.length - 1; i > 0; i--) {
				const j = Math.floor(Math.random() * (i + 1));
				[array[i], array[j]] = [array[j], array[i]];
			}
			return array;
		},
		
		// 移除或简化selectSubject方法，因为我们只关注AI学科
		async selectSubject(subject) {
			// 如果需要保留切换功能，可以保留这个方法
			// 但是只在选择的是AI学科时加载论文
			if (subject.id === this.aiSubjectId) {
				await this.loadAIPapers();
			}
		},
		
		// 获取引用作者
		getCitation(citation) {
			if (!citation) return 'Unknown Author';
			return citation.split(',')[0];
		},
		
		// 获取年份
		getYear(date) {
			if (!date) return '';
			return new Date(date).getFullYear();
		},
		
		// 获取通过分数
		getPassScore(passCondition) {
			try {
				if (!passCondition) return '未设置';
				const condition = JSON.parse(passCondition);
				return condition.min_score || '未设置';
			} catch (error) {
				return '未设置';
			}
		},
		
		// 打开论文详情
		async openPaper(paper) {
			try {
				if (!paper.level) {
					uni.showToast({
						title: '该论文暂无关卡',
						icon: 'none'
					});
					return;
				}
				
				this.loading = true;
				this.currentPaper = paper;
				
				// 获取关卡问题列表
				const response = await getLevelQuestions(paper.level.id);
				if (response.success) {
					this.questions = response.data.map(question => ({
						...question,
						content: JSON.parse(question.content_json || '{}'),
						answer: JSON.parse(question.answer_json || '{}')
					}));
				}
			} catch (error) {
				this.handleApiError(error);
			} finally {
				this.loading = false;
			}
		},
		
		// 获取问题类型显示文本
		getQuestionType(type) {
			const types = {
				'mcq': '选择题',
				'essay': '简答题'
			};
			return types[type] || '未知类型';
		},
		
		// 获取难度显示文本
		getDifficulty(level) {
			const levels = {
				1: '入门',
				2: '简单',
				3: '中等',
				4: '困难',
				5: '专家'
			};
			return levels[level] || '未知难度';
		},
		
		// 开始答题
		startQuiz(question) {
			// 从问题对象中获取 paper_id
			const paperId = question.paper_id;
			
			if (paperId) {
				uni.navigateTo({
					url: `/pages/quiz/quiz?id=${paperId}`
				});
			} else {
				uni.showToast({
					title: '无法获取论文信息',
					icon: 'none'
				});
			}
		},
		
		navigateTo(page) {
			const routes = {
				map: '/pages/map/map',
				daily: '/pages/daily/daily',
				achievement: '/pages/goal/goal',
				profile: '/pages/profile/profile'
			};
			
			if (page !== 'daily') {
				uni.navigateTo({
					url: routes[page]
				});
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
	font-size: 72rpx;
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

.daily-intro {
	text-align: center;
	margin-bottom: 60rpx;
}

.intro-title {
	display: block;
	font-family: 'PingFang';
	font-size: 56rpx;
	color: #3e2a1c;
	font-weight: bold;
	margin-bottom: 20rpx;
}

.intro-subtitle {
	display: block;
	font-size: 32rpx;
	color: #666666;
}

.subject-selector {
	display: flex;
	justify-content: center;
	gap: 30rpx;
	margin-top: 20rpx;
}

.subject-item {
	padding: 10rpx 20rpx;
	border-radius: 30rpx;
	background-color: #f0f0f0;
	color: #333;
	font-size: 28rpx;
	font-weight: bold;
	cursor: pointer;
	transition: background-color 0.3s ease;
}

.subject-item.active {
	background-color: #ff7f50;
	color: #fff;
}

.subject-item:active {
	opacity: 0.8;
}

/* 论文列表 */
.paper-cards {
	display: flex;
	flex-direction: column;
	gap: 30rpx;
	padding: 20rpx;
}

.paper-card {
	background: #ffffff;
	border-radius: 20rpx;
	padding: 30rpx;
	box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.1);
	transition: transform 0.3s ease;
}

.paper-card:active {
	transform: scale(0.98);
}

.card-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 20rpx;
}

.paper-number {
	font-size: 32rpx;
	font-weight: bold;
	color: #ff7f50;
}

.paper-category {
	font-size: 24rpx;
	color: #666;
}

.card-title {
	font-size: 32rpx;
	color: #333;
	line-height: 1.5;
	margin-bottom: 20rpx;
	font-weight: 500;
}

.card-meta {
	display: flex;
	justify-content: space-between;
	font-size: 24rpx;
	color: #666;
	margin-bottom: 20rpx;
}

.card-description {
	font-size: 24rpx;
	color: #666;
	line-height: 1.4;
	margin-bottom: 20rpx;
}

.level-info {
	margin-top: 20rpx;
	padding-top: 20rpx;
	border-top: 2rpx solid #f0f0f0;
	display: flex;
	justify-content: space-between;
	align-items: center;
}

.level-name {
	font-size: 28rpx;
	color: #3e2a1c;
	font-weight: bold;
}

.level-score {
	font-size: 24rpx;
	color: #666;
}

.author, .year {
	font-size: 24rpx;
	color: #978B6B;
}

/* 问题列表样式 */
.question-section {
	padding: 20rpx;
}

/* 修改问题卡片样式 */
.question-cards {
	display: flex;
	flex-direction: column;
	gap: 30rpx;
	padding: 20rpx;
}

.question-card {
	background: #ffffff;
	border-radius: 20rpx;
	padding: 30rpx;
	box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.1);
	transition: transform 0.3s ease;
}

.question-card:active {
	transform: scale(0.98);
}

.card-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 20rpx;
}

.question-number {
	font-size: 32rpx;
	font-weight: bold;
	color: #ff7f50;
}

.question-meta {
	display: flex;
	gap: 16rpx;
}

.question-type,
.question-difficulty {
	font-size: 24rpx;
	padding: 4rpx 16rpx;
	border-radius: 16rpx;
	background: #f5f5f5;
	color: #666;
}

.card-title {
	font-size: 32rpx;
	color: #333;
	line-height: 1.5;
	margin-bottom: 20rpx;
	font-weight: 500;
}

.card-meta {
	display: flex;
	justify-content: space-between;
	font-size: 24rpx;
	color: #666;
	margin-bottom: 20rpx;
}

.card-preview {
	margin-top: 20rpx;
	padding-top: 20rpx;
	border-top: 2rpx solid #f0f0f0;
}

.preview-title {
	font-size: 24rpx;
	color: #999;
	margin-bottom: 10rpx;
}

.preview-options {
	display: flex;
	gap: 20rpx;
}

.preview-option {
	width: 40rpx;
	height: 40rpx;
	background: #f8f8f8;
	border-radius: 20rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 24rpx;
	color: #666;
}

/* 根据难度设置不同的边框颜色 */
.question-card[data-difficulty="1"] {
	border-left: 8rpx solid #4CAF50;
}

.question-card[data-difficulty="2"] {
	border-left: 8rpx solid #2196F3;
}

.question-card[data-difficulty="3"] {
	border-left: 8rpx solid #FF9800;
}

.question-card[data-difficulty="4"] {
	border-left: 8rpx solid #F44336;
}

.question-card[data-difficulty="5"] {
	border-left: 8rpx solid #9C27B0;
}

.back-button {
	display: flex;
	align-items: center;
	padding: 20rpx;
	color: #666;
	font-size: 28rpx;
}

.back-icon {
	margin-right: 10rpx;
	font-size: 32rpx;
}

.current-paper-info {
	padding: 30rpx;
	background: #fff;
	margin-bottom: 20rpx;
	border-radius: 20rpx;
}

.paper-title {
	font-size: 32rpx;
	font-weight: bold;
	color: #333;
	margin-bottom: 10rpx;
	display: block;
}

.paper-author {
	font-size: 24rpx;
	color: #666;
}

/* 添加论文信息样式 */
.paper-info {
	background: #f8f9fa;
	padding: 16rpx;
	border-radius: 12rpx;
	margin-bottom: 20rpx;
}

.paper-title {
	font-size: 28rpx;
	color: #333;
	font-weight: bold;
	display: block;
	margin-bottom: 8rpx;
}

.paper-author {
	font-size: 24rpx;
	color: #666;
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

/* 确保内容区域不被底部栏遮挡 */
.content-area {
    padding-bottom: 240rpx;
}

/* 骨骼屏动画样式 */
.skeleton-cards {
  display: flex;
  flex-direction: column;
  gap: 30rpx;
  padding: 20rpx;
}

.skeleton-card {
  background: #f5f5f5;
  border-radius: 20rpx;
  padding: 30rpx;
  box-shadow: 0 4rpx 12rpx rgba(0,0,0,0.05);
  animation: skeleton-loading 1.2s infinite linear alternate;
}

@keyframes skeleton-loading {
  0% { background-color: #f5f5f5; }
  100% { background-color: #e9ecef; }
}

.skeleton-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
}

.skeleton-number {
  width: 80rpx;
  height: 32rpx;
  border-radius: 8rpx;
  background: #e0e0e0;
}

.skeleton-meta {
  display: flex;
  gap: 16rpx;
}

.skeleton-tag {
  width: 60rpx;
  height: 24rpx;
  border-radius: 12rpx;
  background: #e0e0e0;
}

.skeleton-paper-info {
  margin-bottom: 20rpx;
}

.skeleton-paper-title {
  width: 180rpx;
  height: 28rpx;
  border-radius: 8rpx;
  background: #e0e0e0;
  margin-bottom: 8rpx;
}

.skeleton-paper-author {
  width: 120rpx;
  height: 20rpx;
  border-radius: 8rpx;
  background: #e0e0e0;
}

.skeleton-title {
  width: 90%;
  height: 32rpx;
  border-radius: 8rpx;
  background: #e0e0e0;
  margin-bottom: 12rpx;
}
.skeleton-title.short {
  width: 60%;
}

.skeleton-score {
  width: 80rpx;
  height: 20rpx;
  border-radius: 8rpx;
  background: #e0e0e0;
}

.skeleton-preview {
  margin-top: 20rpx;
}

.skeleton-preview-title {
  width: 100rpx;
  height: 20rpx;
  border-radius: 8rpx;
  background: #e0e0e0;
  margin-bottom: 10rpx;
}

.skeleton-options {
  display: flex;
  gap: 20rpx;
}

.skeleton-option {
  width: 40rpx;
  height: 40rpx;
  border-radius: 20rpx;
  background: #e0e0e0;
}
</style>