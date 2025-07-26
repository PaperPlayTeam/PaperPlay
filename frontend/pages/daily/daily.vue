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
				<!-- 学科选择器 -->
				<view class="subject-selector">
					<view 
						v-for="subject in subjects" 
						:key="subject.id"
						:class="['subject-item', { active: currentSubject && currentSubject.id === subject.id }]"
						@click="selectSubject(subject)"
					>
						{{ subject.name }}
					</view>
				</view>
			</view>

			<!-- 论文列表 -->
			<view v-if="loading" class="loading-container">
				<text>加载中...</text>
			</view>
			<view v-else-if="!currentPaper" class="paper-cards">
				<view 
					v-for="(paper, index) in papers" 
					:key="paper.id"
					class="paper-card"
					@click="openPaper(paper)"
				>
					<view class="card-header">
						<text class="paper-number">#{{ index + 1 }}</text>
						<text class="paper-category" v-if="currentSubject">{{ currentSubject.name }}</text>
					</view>
					<view class="card-title">
						<text>{{ paper.title }}</text>
					</view>
					<view class="card-meta">
						<text class="author">{{ getCitation(paper.citation) }}</text>
						<text class="year">{{ getYear(paper.created_at) }}</text>
					</view>
					<view class="card-description">
						<text>{{ paper.description || '探索这篇论文的核心概念和关键发现。' }}</text>
					</view>
					<view v-if="paper.level" class="level-info">
						<text class="level-name">关卡: {{ paper.level.name }}</text>
						<text class="level-score">通过分数: {{ getPassScore(paper.level.pass_condition) }}</text>
					</view>
				</view>
			</view>
			
			<!-- 问题列表 -->
			<view v-else class="question-section">
				<!-- 返回按钮 -->
				<view class="back-button" @click="currentPaper = null">
					<text class="back-icon">←</text>
					<text>返回论文列表</text>
				</view>
				
				<!-- 当前论文信息 -->
				<view class="current-paper-info">
					<text class="paper-title">{{ currentPaper.title }}</text>
					<text class="paper-author">{{ getCitation(currentPaper.citation) }}</text>
				</view>
				
				<!-- 问题卡片列表 -->
				<view class="question-cards">
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
								<text class="question-type">{{ getQuestionType(question.content.type) }}</text>
								<text class="question-difficulty">{{ getDifficulty(question.difficulty) }}</text>
							</view>
						</view>
						<view class="card-title">
							<text>{{ question.stem }}</text>
						</view>
						<view class="card-meta">
							<text class="score">分值: {{ question.score }}分</text>
							<text class="author">出题人: {{ question.created_by }}</text>
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
			subjects: [],
			currentSubject: null,
			papers: [],
			loading: false,
			currentPaper: null,
			questions: []
		}
	},
	async onLoad() {
		await this.loadSubjects();
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
		
		// 加载所有学科
		async loadSubjects() {
			try {
				this.loading = true;
				const response = await getAllSubjects();
				if (response.success && response.data.length > 0) {
					this.subjects = response.data;
					// 默认选择第一个学科
					await this.selectSubject(this.subjects[0]);
				}
			} catch (error) {
				this.handleApiError(error);
			} finally {
				this.loading = false;
			}
		},
		
		// 选择学科
		async selectSubject(subject) {
			try {
				this.loading = true;
				this.currentSubject = subject;
				this.questions = []; // 清空问题列表
				this.currentPaper = null;
				
				const response = await getSubjectPapers(subject.id);
				if (response.success) {
					// 获取每个论文的关卡信息
					const papersWithLevels = await Promise.all(
						response.data.map(async (paper) => {
							try {
								const levelResponse = await getPaperLevel(paper.id);
								return {
									...paper,
									level: levelResponse.success ? levelResponse.data : null
								};
							} catch (error) {
								console.error(`获取论文${paper.id}的关卡信息失败:`, error);
								return paper;
							}
						})
					);
					this.papers = papersWithLevels;
				}
			} catch (error) {
				this.handleApiError(error);
			} finally {
				this.loading = false;
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
			if (this.currentPaper && this.currentPaper.level) {
				uni.navigateTo({
					url: `/pages/quiz/quiz?levelId=${this.currentPaper.level.id}&questionId=${question.id}`
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

.question-cards {
	display: flex;
	flex-direction: column;
	gap: 20rpx;
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
</style>