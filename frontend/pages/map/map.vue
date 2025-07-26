<template>
	<view class="container">
		<view class="header" style="z-index: 10000;">
			<view class="title-container">
				<text class="title">{{ fieldName }} <br></text>
				<text class="fieldDescription">{{fieldDescription}}</text>
			</view>
		</view>
		<view id="map" class="map-container"></view>
		
		<!-- 灰色蒙版 -->
		<view 
			class="mask" 
			:class="{ 'mask-show': isTabShow }"
			@click="closeTab"
		></view>
		
		<!-- 左侧tab栏 -->
		<view class="side-tab" :class="{ 'side-tab-show': isTabShow }">
			<view class="tab-content">
				<text class="tab-title">click to change<br>the map</text>
				<view class="map-icons">
					<image src="/static/area/ai.svg" class="map-icon" mode="aspectFit"></image>
					<image src="/static/area/biology.svg" class="map-icon" mode="aspectFit"></image>
					<image src="/static/area/math.svg" class="map-icon" mode="aspectFit"></image>
				</view>
			</view>
		</view>
		
		<!-- 底部背景 -->
		<view class="bottom-bg">
			<image src="/static/docker/梯形.svg" mode="aspectFill"></image>
		</view>
		
		<!-- 底部装饰栏 -->
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
			<view class="trapezoid t4" @click="navigateTo('profile')">
				<image style="z-index: 20;margin-left: 40rpx;" src="/static/docker/我.svg" class="docker-icon" mode="aspectFit"></image>
			</view>
		</view>
		
		<image 
			:style="{
				width: '480rpx',
				height: '480rpx',
				position: 'fixed',
				left: `${mascotLeft}px`,
				bottom: '60rpx',
				transition: isDragging ? 'none' : 'left 0.3s ease-out'
			}"
			:src="mascotSrc" 
			class="mascot" 
			mode="aspectFill"
			@touchstart="handleTouchStart"
			@touchmove="handleTouchMove"
			@touchend="handleTouchEnd"
			@click="toggleMascot"
		></image>
		<!-- <image src="/static/talk.svg" style="z-index: 1000;width:400rpx;height:200rpx;position:fixed;left: 330rpx;bottom: 120rpx;" mode="scaleToFill"></image> -->
		<!-- <view class="wrapper">
			<svg class="bottom-arch" viewBox="0 0 100 10" preserveAspectRatio="none">
  <path d="M0,0 C25,3 75,5 100,0 L100,10 L0,10 Z" fill="#000"/>
</svg>

  </view> -->
	</view>
</template>		

<script>
import 'leaflet/dist/leaflet.css';
import L from 'leaflet';
import rough from 'roughjs';
import { getAllSubjects, getSubjectPapers } from '@/api/subjects';

export default {
	data() {
		return {
			map: null,
			roughInstance: null,
			mascotSrc: '/static/cute.png',
			isAnimating: false,
			isDrawMode: false,
			connections: [],
			fieldDescription:'',
			fieldName:'',
			isTabShow: false,
			touchStartX: 0,
			touchStartY: 0,
			isDragging: false,
			mascotLeft: -80, // 初始位置
			loading: false,
			nodes: []
		}
	},
	mounted() {
		// 预加载静态图片
		const preloadJpg = new Image();
		preloadJpg.src = '/static/cute.png';
		preloadJpg.onload = () => {
			this.mascotSrc = '/static/cute.png';
		};

		// 初始化 Rough.js
		this.roughInstance = rough;
		
		// 等待DOM加载完成后初始化
		this.$nextTick(async () => {
			await this.loadMapData();
			this.initMap();
		});
	},
	methods: {
		closeTab() {
			this.isTabShow = false;
			this.mascotLeft = -80;
		},
		handleTouchStart(event) {
			this.touchStartX = event.touches[0].clientX;
			this.touchStartY = event.touches[0].clientY;
			this.isDragging = true;
		},
		handleTouchMove(event) {
			if (!this.isDragging) return;
			
			const deltaX = event.touches[0].clientX - this.touchStartX;
			const deltaY = Math.abs(event.touches[0].clientY - this.touchStartY);
			
			// 如果是横向拖动
			if (deltaX > deltaY) {
				// 如果tab栏已经显示，不允许拖动
				if (this.isTabShow) return;
				
				// 限制最大拖动距离
				const newLeft = Math.min(Math.max(-80, -80 + deltaX), 180);
				this.mascotLeft = newLeft;
				
				// 当向右拖动超过阈值时显示tab栏
				if (deltaX > 30) {
					this.isTabShow = true;
				}
				
				event.preventDefault(); // 阻止页面滚动
			}
		},
		handleTouchEnd() {

			this.isDragging = false;
			
			// 根据tab栏状态决定最终位置
			if (!this.isTabShow) {
				this.mascotLeft = -80; // 回到初始位置
			} else {
				this.mascotLeft = 180;  // 移动到最终位置
			}
			
			this.touchStartX = 0;
			this.touchStartY = 0;
		},
		navigatorToDaily(){
			uni.navigateTo({
	url: 'pages/daily/daily'
});
		},
		async loadMapData() {
			try {
				this.loading = true;
				
				// 预定义的坐标数组
				const coordinates = [
					{x: 171, y: 338},
					{x: 124, y: 348},
					{x: 103, y: 406},
					{x: 86, y: 547},
					{x: 113, y: 644},
					{x: 113, y: 644},
					{x: 246, y: 551},
					{x: 338, y: 503},
					{x: 363, y: 536},
					{x: 407, y: 604}
				];
				
				// 获取所有学科
				const subjectsResponse = await getAllSubjects();
				if (!subjectsResponse.success || !subjectsResponse.data.length) {
					throw new Error('没有找到学科数据');
				}
				this.fieldName = subjectsResponse.data[2].name;
				this.fieldDescription = subjectsResponse.data[2].description;

				// 获取第一个学科的论文
				const firstSubject = subjectsResponse.data[2];
				const papersResponse = await getSubjectPapers(firstSubject.id);
				
				if (!papersResponse.success) {
					throw new Error('获取论文数据失败');
				}

				// 转换论文数据为节点格式
				this.nodes = papersResponse.data.map((paper, index) => {
					// 获取预定义坐标，如果超出数组范围则使用最后一个坐标
					const coord = coordinates[index] || coordinates[coordinates.length - 1];
					
					// 从citation中提取作者
					const author = paper.citation ? paper.citation.split(',')[0] : 'Unknown Author';
					// 从created_at中提取年份
					const year = paper.created_at ? new Date(paper.created_at).getFullYear() : new Date().getFullYear();
					
					return {
						id: paper.id,
						title: paper.title,
						author: author,
						year: year,
						keywords: firstSubject.name,
						citations: 0,
						x: coord.x,
						y: coord.y,
						zone: index % 3 === 0 ? 'high' : (index % 3 === 1 ? 'medium' : 'low'),
						unlocked: false,
						parent_id: index > 0 ? papersResponse.data[index - 1].id : null,
						sort_order: index + 1
					}
				});
			} catch (error) {
				console.error('加载数据失败:', error);
				uni.showToast({
					title: error.message || '加载数据失败',
					icon: 'none'
				});
			} finally {
				this.loading = false;
			}
		},
		initMap() {
			try {
				// 创建地图实例
				this.map = L.map('map', {
					crs: L.CRS.Simple,
					minZoom: -2,  // 设置最小缩放级别
					maxZoom: 2,  // 设置最大缩放级别与最小相同，禁用缩放
					zoomControl: false,  // 移除缩放控件
					attributionControl: false,
					dragging: true  // 保持可拖动
				});

				// 设置边界
				const bounds = [[0, 0], [1000, 1000]];
				this.map.setMaxBounds(bounds);
				
				// 设置初始视图
				this.map.setView([700, 300], 0);

				// 添加背景
				L.imageOverlay('/static/math.svg', bounds).addTo(this.map);

				// 添加地图点击事件
				this.map.on('click', (e) => {
					// 获取点击位置的坐标
					const y = Math.round(1000 - e.latlng.lat); // 转换为正常坐标系
					const x = Math.round(e.latlng.lng);
					
					console.log('点击坐标:', { x, y });
					
			
					
					// 如果需要，可以在这里添加新的节点
					// this.addNode(x, y);
				});

				// 先创建所有节点
				this.nodes.forEach(node => {
					this.createNode(node);
				});

			
			} catch (error) {
				console.error('初始化地图失败:', error);
			}
		},
		createNode(data) {
			try {
				const mainSvg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
				mainSvg.setAttribute('width', '100');
				mainSvg.setAttribute('height', '100');
				mainSvg.setAttribute('viewBox', '0 0 100 100');
				const rc = this.roughInstance.svg(mainSvg);

				// 创建外圆（透明）
				const outerCircle = rc.circle(50, 50, 15, {
					stroke: '#000',
					strokeWidth: 2,
					roughness: 1.5,
					fill: 'none',
					bowing: 1
				});
				mainSvg.appendChild(outerCircle);

				// 创建内部小圆点
				const innerCircle = rc.circle(50, 50, 3, {
					stroke: '#000',
					strokeWidth: 1,
					roughness: 1,
					fill: '#000',
					fillStyle: 'solid',
					bowing: 1
				});
				mainSvg.appendChild(innerCircle);

				// 添加文字
				const text = document.createElementNS("http://www.w3.org/2000/svg", "text");
				text.setAttribute('x', 50);
				text.setAttribute('y', 75);
				text.setAttribute('text-anchor', 'middle');
				text.setAttribute('dominant-baseline', 'hanging');
				text.setAttribute('font-family', 'PingFang');
				text.setAttribute('font-size', '24rpx');
				text.setAttribute('fill', '#3e2a1c');
				// text.setAttribute('style', 'white-space: nowrap;');
				text.textContent = data.title;
				mainSvg.appendChild(text);

				const icon = L.divIcon({
					className: '',
					html: `<div class="node-container">${mainSvg.outerHTML}</div>`,
					iconSize: [100, 100],
					iconAnchor: [50, 50]
				});
				
				const y = 1000 - Math.max(0, Math.min(data.y, 1000));
				const x = Math.max(0, Math.min(data.x, 1000));
				
				const marker = L.marker([y, x], {
					icon: icon,
					title: data.title
				});

				marker.on('click', () => {
					console.log('点击marker，id=', data.id);
					uni.navigateTo({
						url: `/pages/quiz/quiz?id=${data.id}`,
						fail: (err) => {
							console.error('跳转失败:', err);
							uni.showToast({
								title: '跳转失败',
								icon: 'none'
							});
						}
					});
				});
				
				marker.addTo(this.map);
			} catch (error) {
				console.error('创建节点失败:', error);
			}
		},
		toggleMascot() {
			// 切换模式
			this.isDrawMode = !this.isDrawMode;
			
			if (this.isDrawMode) {
				// 切换到绘制模式
				this.map.setMinZoom(0);
				this.map.setMaxZoom(0);
				this.map.setZoom(0, { animate: false });
				if (this.map.zoomControl) {
					this.map.zoomControl.remove();
				}
				// 绘制连线
				this.drawAllConnections();
				
				
				// 播放GIF动画
				const preloadGif = new Image();
				preloadGif.src = '/static/cute.GIF';
				
				preloadGif.onload = () => {
					if (!this.isAnimating) {
						this.isAnimating = true;
						const staticImage = '/static/cute.png';
						const animatedImage = '/static/cute.GIF';
						
						this.mascotSrc = animatedImage;
						
						setTimeout(() => {
							this.mascotSrc = staticImage;
							this.isAnimating = false;
						}, 3000);
					}
				};
				
				preloadGif.onerror = (error) => {
					console.error('GIF加载失败:', error);
					this.mascotSrc = '/static/cute.png';
					this.isAnimating = false;
				};
			} else {
				// 切换到自由查看模式
				this.map.setMinZoom(-1);
				this.map.setMaxZoom(2);
				// L.control.zoom({
				// 	position: 'bottomright'
				// }).addTo(this.map);
				// 清除连线
				this.clearConnections();
				
				// 播放GIF动画
				const preloadGif = new Image();
				preloadGif.src = '/static/cute.GIF';
				
				preloadGif.onload = () => {
					if (!this.isAnimating) {
						this.isAnimating = true;
						const staticImage = '/static/cute.png';
						const animatedImage = '/static/cute.GIF';
						
						this.mascotSrc = animatedImage;
						
						setTimeout(() => {
							this.mascotSrc = staticImage;
							this.isAnimating = false;
						}, 3000);
					}
				};
				
				preloadGif.onerror = (error) => {
					console.error('GIF加载失败:', error);
					this.mascotSrc = '/static/cute.png';
					this.isAnimating = false;
				};
			}
		},
		drawAllConnections() {
			// 先清除现有连线
			this.clearConnections();
			
			// 绘制所有连线
			this.nodes.forEach(node => {
				if (node.parent_id) {
					this.drawConnection(node);
				}
			});
		},
		clearConnections() {
			// 清除所有连线
			this.connections.forEach(connection => {
				connection.remove();
			});
			this.connections = [];
		},
		drawConnection(node) {
			try {
				const parentNode = this.nodes.find(n => n.id === node.parent_id);
				if (!parentNode) return;

				// 创建SVG路径
				const mainSvg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
				mainSvg.setAttribute('width', '1000');
				mainSvg.setAttribute('height', '1000');
				mainSvg.style.position = 'absolute';
				mainSvg.style.top = '0';
				mainSvg.style.left = '0';
				mainSvg.style.pointerEvents = 'none';
				mainSvg.style.zIndex = '999';

				const rc = this.roughInstance.svg(mainSvg);
				
				// 计算连线坐标
				const line = rc.line(
					node.x,
					node.y,
					parentNode.x,
					parentNode.y,
					{
						stroke: '#978B6B',
						strokeWidth: 2,
						roughness: 1
					}
				);
				mainSvg.appendChild(line);

				// 将SVG添加到地图
				const overlay = L.svgOverlay(mainSvg, [[0, 0], [1000, 1000]], {
					interactive: false,
					className: 'connection-line'
				});
				overlay.addTo(this.map);
				
				// 保存连线引用
				this.connections.push(overlay);
			} catch (error) {
				console.error('绘制连线失败:', error);
			}
		},
		navigateTo(page) {
			const routes = {
				map: '/pages/map/map',
				daily: '/pages/daily/daily',
				achievement: '/pages/goal/goal',
				profile: '/pages/profile/profile'
			};
			
			if (page !== 'map') {
				uni.navigateTo({
					url: routes[page]
				});
			}
		}
	}
}</script>

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
    position: relative;
}

.header {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    height: 180rpx;
    padding: 90rpx 40rpx 20rpx;
    display: flex;
    justify-content: space-between;
    align-items: center;
    z-index: 10;
    pointer-events: none;
}

.title-container {
    position: absolute;
    left: 50%;
    transform: translateX(-50%);
    text-align: center;
}

.title {
    font-family: 'Fredoka One', 'PingFang', cursive;
    font-size: 56rpx;
    color: #000;
    letter-spacing: 2rpx;
    text-shadow: 4rpx 4rpx 0rpx rgba(255,255,255,0.5);
}
.fieldDescription{
	font-family: 'PingFang';
    font-size: 24rpx;
    color: #000;
    letter-spacing: 2rpx;
    text-shadow: 4rpx 4rpx 0rpx rgba(255,255,255,0.5);
}
.icons {
    position: absolute;
    right: 40rpx;
    display: flex;
    flex-direction: column;
    gap: 36rpx;
    pointer-events: auto;
}

.icon {
    width: 92rpx;
    height: 92rpx;
    box-shadow: 0 12rpx 32rpx -8rpx rgba(0,0,0,0.18);
    border-radius: 24rpx;
    transition: box-shadow 0.2s;
}

.icon:active {
    box-shadow: 0 4rpx 12rpx rgba(0,0,0,0.12);
}

.map-container {
    width: 100%;
    height: 100%;
    background: #ADD7F5;
}

/* 节点样式 */
.node-container {
	position: relative;
	height: 40px;
	width: 300px;
}

.node-title {
    position: absolute;
    bottom: 50rpx;
    left: 160rpx;
    padding: 8rpx 0;
    width: 400rpx;
    height: 80rpx;
}

.node-title svg {
    width: 100%;
    height: 100%;
}

.title-text {
    margin-left: 10rpx;
    font-family: PingFang !important;
    color: #978B6B;
    font-size: 48rpx;
    text-shadow: 4rpx 4rpx 8rpx rgba(255,255,255,0.8);
}

.node-marker {
	position: absolute;
	left: 30rpx;
	top: 30rpx;
	width: 32rpx;
	height: 32rpx;
}

.node-marker svg {
	width: 100%;
	height: 100%;
}

.connector svg {
	position: absolute;
	left: 62rpx;
	top: 20rpx;
	width: 80rpx;
	height: 60rpx;
}

.paper-node {
    background: white;
    border: 4rpx solid;
    padding: 20rpx;
    border-radius: 8rpx;
    min-width: 400rpx;
    font-family: PingFang !important;
}

.paper-node.high {
	border-color: #b77a56;
}

.paper-node.medium {
	border-color: #e7c190;
}

.paper-node.low {
	border-color: #a2c69b;
}

.paper-node .title {
    font-family: PingFang !important;
    font-weight: bold;
    margin-bottom: 16rpx;
    font-size: 28rpx;
    color: #978B6B;
}

.paper-node .info {
    font-family: PingFang !important;
    font-size: 24rpx;
    line-height: 1.4;
    color: #978B6B;
}

.paper-node .citations {
    font-family: PingFang !important;
    margin-top: 16rpx;
    padding-top: 16rpx;
    border-top: 2rpx solid #eee;
    font-weight: bold;
    color: #978B6B;
}

.node-container {
    position: relative;
    width: 200rpx;
    height: 200rpx;
    overflow: visible; /* 允许内容超出容器 */
}

.node-container svg {
    width: 100%;
    height: 100%;
    position: absolute;
    left: 0;
    top: 0;
    overflow: visible; /* 允许SVG内容超出边界 */
}

text {
    white-space: nowrap !important;
    overflow: visible !important;
}

.mascot {
	z-index: 9999;
	cursor: pointer;
	will-change: transform; /* 优化动画性能 */
}

.node-info-modal {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;
}

.node-info-content {
    background: white;
    border-radius: 40rpx;
    width: 80%;
    max-width: 1000rpx;
    padding: 40rpx;
    box-shadow: 0 8rpx 40rpx rgba(0, 0, 0, 0.15);
}

.node-info-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 40rpx;
}

.node-info-title {
    font-size: 40rpx;
    font-weight: bold;
    color: #3e2a1c;
}

.node-info-close {
    font-size: 48rpx;
    padding: 20rpx;
    cursor: pointer;
}

.node-info-body {
    margin-bottom: 40rpx;
}

.node-info-meta {
    display: flex;
    flex-direction: column;
    gap: 16rpx;
    margin-bottom: 30rpx;
    color: #666;
}

.node-info-status {
    margin: 20rpx 0;
    padding: 10rpx 20rpx;
    border-radius: 8rpx;
    display: inline-block;
}

.node-info-status.locked {
    background: #ffebee;
    color: #d32f2f;
}

.node-info-status.unlocked {
    background: #e8f5e9;
    color: #2e7d32;
}

.node-info-description {
    margin-top: 30rpx;
    line-height: 1.6;
    color: #333;
}

.node-info-footer {
    text-align: center;
    margin-top: 40rpx;
}

.node-info-button {
    padding: 24rpx 60rpx;
    border-radius: 50rpx;
    border: none;
    font-size: 32rpx;
    cursor: pointer;
    transition: all 0.3s;
}

.node-info-button.primary {
    background: #3e2a1c;
    color: white;
}

.node-info-button.primary:active {
    background: #2a1c0f;
}

.node-info-button.disabled {
    background: #ccc;
    color: #666;
    cursor: not-allowed;
}
.bottom-container {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  z-index: 10;
}
.bottom-arch {
  position: fixed;   /* 或 absolute，根据你需求 */
  left: 0;
  bottom: 0;
  width: 100%;
  height: 140rpx;     /* 调这行控制弯的可见高度 */
  z-index: 999;
}

.side-tab {
	position: fixed;
	left: -600rpx;
	bottom: 0;
	width: 560rpx;
	height: 70vh;
	padding-left: 40rpx;
	background: #272828;
	box-shadow: 0 0 40rpx rgba(0, 0, 0, 0.1);
	transition: transform 0.3s ease-out;
	border-radius: 40rpx 40rpx 0 0;
	z-index: 9998;
}

.side-tab-show {
	transform: translateX(600rpx);
}

.tab-content {
	padding: 80rpx 40rpx;
}

.tab-title {
	font-family: 'PingFang';
	font-size: 56rpx;
	color: #ffffff;
	margin-bottom: 80rpx;
	display: block;
	text-align: center;
}

.map-icons {
	display: flex;
	flex-direction: column;
	align-items: flex-start;
	gap: 60rpx;
	padding-left: 80rpx;
}

.map-icon {
	width: 320rpx;
	height: 80rpx;
	border-radius: 24rpx;
	transition: transform 0.2s ease;
}

.map-icon:active {
	transform: scale(0.98);
}

.mask {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background: rgba(0, 0, 0, 0.5);
	opacity: 0;
	visibility: hidden;
	transition: all 0.3s ease-out;
	z-index: 9997;
}

.mask-show {
	opacity: 1;
	visibility: visible;
}

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
	
	margin-left: -40rpx; /* 创建重叠效果 */
}

.trapezoid:first-child {
	margin-left: 0;
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

.trapezoid::before {
	content: '';
	position: absolute;
	bottom: 0;
	left: 0;
	right: 0;
	height: 100%;
}

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

</style>