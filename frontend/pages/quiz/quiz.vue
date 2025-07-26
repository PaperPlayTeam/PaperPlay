<template>
	<view class="quiz-container">
		<!-- 顶部返回按钮 -->
		<view class="quiz-header">
			<view class="back-btn" @click="goBack">
				<text class="back-icon">←</text>
				<text>返回</text>
			</view>
		</view>

		<!-- 骨骼屏动画 -->
		<view v-if="loading" class="quiz-skeleton">
			<view class="skeleton-title"></view>
			<view class="skeleton-card"></view>
			<view class="skeleton-card"></view>
			<view class="skeleton-btn"></view>
		</view>

		<view v-else>
			<!-- 当前题对的标题 -->
			<view class="concept-title">
				{{ currentConceptName }}
			</view>

			<!-- 引入题部分 -->
			<view class="question-section">
				<view class="question-type">引入题</view>
				<view class="question-card" :class="{ 'completed': showConceptQuestion }">
					<view class="question-text">{{ leadInQuestionContent.question }}</view>
					
					<view class="options-container">
						<view v-for="(option, index) in leadInQuestionContent.options" 
							:key="index"
							class="option"
							:class="{
								'selected': leadInSelectedOption === index,
								'correct': leadInShowResult && isLeadInOptionCorrect(index),
								'wrong': leadInShowResult && leadInSelectedOption === index && !isLeadInOptionCorrect(index)
							}"
							@click="selectLeadInOption(index)">
							{{ option }}
						</view>
					</view>

					<view v-if="leadInShowResult" class="feedback-card" 
						:class="{ 'correct': isLeadInAnswerCorrect }">
						<view class="feedback-text">
							{{ leadInAnswerContent.explanation }}
						</view>
					</view>
				</view>
			</view>

			<!-- 概念题部分 -->
			<view v-if="showConceptQuestion" class="question-section">
				<view class="question-type">概念题</view>
				<view class="question-card">
					<!-- 概念解释 -->
					<view class="concept-explanation">
						{{ conceptQuestionContent.concept_explanation }}
					</view>

					<view class="question-text">{{ conceptQuestionContent.question }}</view>
					
					<view class="options-container">
						<view v-for="(option, index) in conceptQuestionContent.options" 
							:key="index"
							class="option"
							:class="{
								'selected': conceptSelectedOption === index,
								'correct': conceptShowResult && isConceptOptionCorrect(index),
								'wrong': conceptShowResult && conceptSelectedOption === index && !isConceptOptionCorrect(index)
							}"
							@click="selectConceptOption(index)">
							{{ option }}
						</view>
					</view>

					<view v-if="conceptShowResult" class="feedback-card" 
						:class="{ 'correct': isConceptAnswerCorrect }">
						<view class="feedback-text">
							{{ conceptAnswerContent.explanation }}
						</view>
					</view>
				</view>
			</view>

			<!-- 下一步按钮 -->
			<button v-if="leadInShowResult && (!showConceptQuestion || conceptShowResult)" 
				class="next-button" 
				@click="handleNext">
				{{ nextButtonText }}
			</button>
		</view>
	</view>
</template>

<script>
import { getPaperLevel } from '@/api/papers'
import { getLevelQuestions, startLevel, submitAnswer, completeLevel } from '@/api/levels'
import { getQuestion } from '@/api/questions'
import confetti from 'canvas-confetti'  // 添加这行

export default {
  data() {
    return {
      paperId: '',
      questionPairs: [], // [[id1, id2], [id3, id4], ...] 存储成对的问题ID
      currentPairIndex: 0,
      leadInQuestion: null, // 当前引入题
      conceptQuestion: null, // 当前概念题
      showConceptQuestion: false,
      leadInSelectedOption: null,
      conceptSelectedOption: null,
      leadInShowResult: false,
      conceptShowResult: false,
      loading: false,
      // 添加计时相关数据
      startTime: null,
      leadInStartTime: null,
      conceptStartTime: null
    }
  },

  computed: {
    currentConceptName() {
      if (!this.leadInQuestion) return ''
      try {
        return JSON.parse(this.leadInQuestion.content_json).concept_name
      } catch (e) {
        return ''
      }
    },

    leadInQuestionContent() {
      if (!this.leadInQuestion) return {}
      try {
        return JSON.parse(this.leadInQuestion.content_json)
      } catch (e) {
        return {}
      }
    },

    leadInAnswerContent() {
      if (!this.leadInQuestion) return {}
      try {
        return JSON.parse(this.leadInQuestion.answer_json)
      } catch (e) {
        return {}
      }
    },

    conceptQuestionContent() {
      if (!this.conceptQuestion) return {}
      try {
        return JSON.parse(this.conceptQuestion.content_json)
      } catch (e) {
        return {}
      }
    },

    conceptAnswerContent() {
      if (!this.conceptQuestion) return {}
      try {
        return JSON.parse(this.conceptQuestion.answer_json)
      } catch (e) {
        return {}
      }
    },

    isLeadInAnswerCorrect() {
      return this.getOptionLetter(this.leadInSelectedOption) === this.leadInAnswerContent.correct_option
    },

    isConceptAnswerCorrect() {
      return this.getOptionLetter(this.conceptSelectedOption) === this.conceptAnswerContent.correct_option
    },

    nextButtonText() {
      if (!this.showConceptQuestion) {
        return this.isLeadInAnswerCorrect ? '查看概念题' : '重新作答'
      }
      return this.currentPairIndex < this.questionPairs.length - 1 ? '下一组题目' : '完成'
    }
  },

  onLoad(options) {
    if (options.id) {
      this.paperId = options.id
      this.startTime = Date.now()
      this.loadLevel()
    }
  },

  methods: {
    goBack() {
      uni.navigateBack();
    },
    async loadLevel() {
      this.loading = true
      try {
        const levelResponse = await getPaperLevel(this.paperId)
        
        if (levelResponse.success) {
          this.level = levelResponse.data
          this.levelId = this.level.id // 保存 levelId
          
          // 开始关卡
          try {
            await startLevel(this.levelId)
            console.log('关卡开始成功')
          } catch (error) {
            console.error('开始关卡失败:', error)
            // 不阻止用户继续，只记录错误
          }
          
          await this.loadQuestionIds(this.level.id)
        }
      } catch (error) {
        console.error('获取关卡失败:', error)
        this.handleApiError(error)
      } finally {
        this.loading = false
      }
    },

    async loadQuestionIds(levelId) {
      try {
        const questionsResponse = await getLevelQuestions(levelId)
        
        if (questionsResponse.success) {
          // 将问题ID两两配对
          const allQuestions = questionsResponse.data
          this.questionPairs = []
          
          for (let i = 0; i < allQuestions.length; i += 2) {
            this.questionPairs.push([
              allQuestions[i].id,
              allQuestions[i + 1].id
            ])
          }
          
          // 加载第一个引入题
          if (this.questionPairs.length > 0) {
            await this.loadCurrentPair()
          }
        }
      } catch (error) {
        console.error('获取问题列表失败:', error)
        this.handleApiError(error)
      }
    },

    async loadCurrentPair() {
      this.loading = true
      try {
        // 加载当前对的两个问题
        const [leadInResponse, conceptResponse] = await Promise.all([
          getQuestion(this.questionPairs[this.currentPairIndex][0]),
          getQuestion(this.questionPairs[this.currentPairIndex][1])
        ])

        if (leadInResponse.success && conceptResponse.success) {
          this.leadInQuestion = leadInResponse.data
          this.conceptQuestion = conceptResponse.data
          this.resetQuestionState()
        }
      } catch (error) {
        console.error('加载题目失败:', error)
        this.handleApiError(error)
      } finally {
        this.loading = false
      }
    },

    resetQuestionState() {
      this.showConceptQuestion = false
      this.leadInSelectedOption = null
      this.conceptSelectedOption = null
      this.leadInShowResult = false
      this.conceptShowResult = false
      // 重置计时器
      this.leadInStartTime = Date.now()
    },

    // 添加这两个新方法
    fireBasicConfetti() {
      confetti({
        particleCount: 100,
        spread: 70,
        origin: { y: 0.6 }
      });
    },

    fireSchoolPride() {
      function fire(particleRatio, opts) {
        confetti({
          ...opts,
          origin: { y: 0.7 },
          particleCount: Math.floor(200 * particleRatio)
        });
      }

      fire(0.25, {
        spread: 26,
        startVelocity: 55,
        origin: { x: 0.2 }
      });

      fire(0.25, {
        spread: 26,
        startVelocity: 55,
        origin: { x: 0.8 }
      });

      setTimeout(() => {
        fire(0.2, {
          spread: 60,
          origin: { x: 0.2 }
        });
        fire(0.2, {
          spread: 60,
          origin: { x: 0.8 }
        });
      }, 150);
    },

    // 修改 selectLeadInOption 方法
    async selectLeadInOption(index) {
      if (this.leadInShowResult) return
      
      this.leadInSelectedOption = index
      this.leadInShowResult = true
      
      const isCorrect = this.isLeadInOptionCorrect(index)
      const durationMs = Date.now() - (this.leadInStartTime || this.startTime)
      
      // 提交答案 - 修改 answer_json 格式
      try {
        const answerJson = {
          type: "multiple_choice",
          selected_option: this.getOptionLetter(index),
          correct_option: this.leadInAnswerContent.correct_option,
          is_correct: isCorrect
        }
        
        await submitAnswer(
          this.levelId,
          this.leadInQuestion.id,
          answerJson,
          durationMs
        )
      } catch (error) {
        console.error('提交答案失败:', error)
        // 不阻止用户继续，只记录错误
      }
      
      if (isCorrect) {
        this.showConceptQuestion = true
        this.conceptStartTime = Date.now()
      } else {
        uni.showToast({ title: '答案错误', icon: 'none' })
      }
    },

    // 修改 selectConceptOption 方法
    async selectConceptOption(index) {
      if (this.conceptShowResult) return
      
      this.conceptSelectedOption = index
      this.conceptShowResult = true
      
      const isCorrect = this.isConceptOptionCorrect(index)
      const durationMs = Date.now() - (this.conceptStartTime || this.startTime)
      
      // 提交答案 - 修改 answer_json 格式
      try {
        const answerJson = {
          type: "multiple_choice",
          selected_option: this.getOptionLetter(index),
          correct_option: this.conceptAnswerContent.correct_option,
          is_correct: isCorrect
        }
        
        await submitAnswer(
          this.levelId,
          this.conceptQuestion.id,
          answerJson,
          durationMs
        )
      } catch (error) {
        console.error('提交答案失败:', error)
        // 不阻止用户继续，只记录错误
      }
      
      if (isCorrect) {
        this.fireBasicConfetti()
      } else {
        uni.showToast({ title: '答案错误', icon: 'none' })
      }
    },

    // 修改 handleNext 方法
    async handleNext() {
      if (!this.showConceptQuestion) {
        // 引入题部分
        if (this.isLeadInAnswerCorrect) {
          this.showConceptQuestion = true
          this.conceptStartTime = Date.now()
        } else {
          this.leadInSelectedOption = null
          this.leadInShowResult = false
        }
      } else if (this.conceptShowResult) {
        // 概念题完成后
        if (this.currentPairIndex < this.questionPairs.length - 1) {
          this.currentPairIndex++
          this.loadCurrentPair()
        } else {
          // 完成所有题目，提交关卡完成
          try {
            await completeLevel(this.levelId)
            console.log('关卡完成提交成功')
          } catch (error) {
            console.error('提交关卡完成失败:', error)
            // 不阻止用户继续，只记录错误
          }
          
          this.fireSchoolPride()
          uni.showToast({
            title: '恭喜完成所有题目！',
            icon: 'success',
            duration: 2000,
            complete: () => {
              setTimeout(() => uni.navigateBack(), 2000)
            }
          })
        }
      }
    },

    getOptionLetter(index) {
      return String.fromCharCode(65 + index)
    },

    isLeadInOptionCorrect(index) {
      return this.getOptionLetter(index) === this.leadInAnswerContent.correct_option
    },

    isConceptOptionCorrect(index) {
      return this.getOptionLetter(index) === this.conceptAnswerContent.correct_option
    },

    handleApiError(error) {
      console.error('API错误:', error)
      if (error.message === '请重新登录') {
        uni.removeStorageSync('access_token')
        uni.removeStorageSync('refresh_token')
        uni.removeStorageSync('token_expires_in')
        uni.removeStorageSync('userInfo')
        
        uni.showToast({
          title: '请重新登录',
          icon: 'none',
          duration: 2000,
          complete: () => {
            setTimeout(() => {
              uni.reLaunch({
                url: '/pages/index/index'
              })
            }, 1000)
          }
        })
      } else {
        uni.showToast({
          title: error.message || '发生错误',
          icon: 'none'
        })
      }
    }
  }
}
</script>

<style>
.quiz-container {
	padding: 30rpx;
	background-color: #f5f5f5;
	min-height: 100vh;
}

.question-card {
	background: #ffffff;
	border-radius: 20rpx;
	padding: 40rpx;
	margin-bottom: 30rpx;
	box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.1);
}

.question-text {
	font-size: 32rpx;
	color: #333;
	line-height: 1.6;
	margin-bottom: 30rpx;
	font-weight: 500;
}

.options-container {
	display: flex;
	flex-direction: column;
	gap: 20rpx;
}

.option {
	padding: 24rpx;
	border-radius: 12rpx;
	background: #f8f9fa;
	font-size: 28rpx;
	color: #495057;
	transition: all 0.3s ease;
}

.option.selected {
	background: #e7f5ff;
	color: #228be6;
}

.option.correct {
	background: #d3f9d8;
	color: #2b8a3e;
}

.option.wrong {
	background: #ffe3e3;
	color: #e03131;
}

.explanation-card {
	background: #fff3bf;
	border-radius: 20rpx;
	padding: 30rpx;
	margin-bottom: 30rpx;
}

.explanation-card.final {
	background: #d3f9d8;
}

.explanation-text {
	font-size: 28rpx;
	color: #495057;
	line-height: 1.6;
}

.concept-section {
	margin-top: 40rpx;
	animation: fadeIn 0.5s ease;
}

@keyframes fadeIn {
	from {
		opacity: 0;
		transform: translateY(20rpx);
	}
	to {
		opacity: 1;
		transform: translateY(0);
	}
}

/* 添加新的样式 */
.concept-name {
  font-size: 36rpx;
  font-weight: bold;
  color: #2c3e50;
  margin-bottom: 20rpx;
  padding-bottom: 20rpx;
  border-bottom: 2rpx solid #eee;
}

.explanation-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #2c3e50;
  margin-bottom: 16rpx;
}

.feedback-card {
  background: #fff3bf;
  border-radius: 20rpx;
  padding: 30rpx;
  margin-bottom: 30rpx;
}

.feedback-card.correct {
  background: #d3f9d8;
}

.feedback-text {
  font-size: 28rpx;
  color: #495057;
  line-height: 1.6;
}

/* 添加新样式 */
.question-type {
  font-size: 32rpx;
  color: #666;
  text-align: center;
  margin-bottom: 20rpx;
  padding: 10rpx;
  background: rgba(255, 255, 255, 0.8);
  border-radius: 10rpx;
}

/* 添加新样式 */
.question-section {
  margin-bottom: 30rpx;
}

.concept-title {
  font-size: 40rpx;
  font-weight: bold;
  color: #2c3e50;
  text-align: center;
  margin-bottom: 30rpx;
  padding: 20rpx;
  background: white;
  border-radius: 16rpx;
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.1);
}

.question-card.completed {
  opacity: 0.8;
}

.concept-explanation {
  font-size: 28rpx;
  color: #666;
  background: #f8f9fa;
  padding: 20rpx;
  border-radius: 12rpx;
  margin-bottom: 20rpx;
  line-height: 1.6;
}

/* 在现有样式的最后添加 */
.quiz-container {
  position: relative;
  z-index: 1;
}

canvas {
  position: fixed !important;
  z-index: 999 !important;
  pointer-events: none;
}

/* 顶部返回按钮样式 */
.quiz-header {
  display: flex;
  align-items: center;
  margin-bottom: 20rpx;
}
.back-btn {
  display: flex;
  align-items: center;
  font-size: 32rpx;
  color: #666;
  padding: 10rpx 20rpx;
  border-radius: 12rpx;
  background: #fff;
  box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.06);
  width: fit-content;
  cursor: pointer;
}
.back-btn:active {
  opacity: 0.7;
}
.back-icon {
  font-size: 36rpx;
  margin-right: 8rpx;
}

/* 骨骼屏动画样式 */
.quiz-skeleton {
  padding: 30rpx 0;
}
.skeleton-title {
  width: 60%;
  height: 48rpx;
  background: #e0e0e0;
  border-radius: 12rpx;
  margin: 0 auto 32rpx auto;
  animation: skeleton-loading 1.2s infinite linear alternate;
}
.skeleton-card {
  width: 100%;
  height: 180rpx;
  background: #e0e0e0;
  border-radius: 20rpx;
  margin-bottom: 30rpx;
  animation: skeleton-loading 1.2s infinite linear alternate;
}
.skeleton-btn {
  width: 40%;
  height: 60rpx;
  background: #e0e0e0;
  border-radius: 30rpx;
  margin: 40rpx auto 0 auto;
  animation: skeleton-loading 1.2s infinite linear alternate;
}
@keyframes skeleton-loading {
  0% { background-color: #e0e0e0; }
  100% { background-color: #f5f5f5; }
}
</style>