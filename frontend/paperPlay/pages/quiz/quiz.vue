<script>
import confetti from 'canvas-confetti'

export default {
  // ... 其他代码保持不变 ...

  methods: {
    // 添加基础庆祝特效
    fireBasicConfetti() {
      confetti({
        particleCount: 100,
        spread: 70,
        origin: { y: 0.6 }
      });
    },

    // 添加完成时的双侧特效
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

    selectLeadInOption(index) {
      if (this.leadInShowResult) return
      this.leadInSelectedOption = index
      this.leadInShowResult = true
      
      const isCorrect = this.isLeadInOptionCorrect(index)
      
      if (isCorrect) {
        this.showConceptQuestion = true
        uni.showToast({ title: '回答正确！', icon: 'success' })
        this.fireBasicConfetti() // 添加正确时的特效
      } else {
        uni.showToast({ title: '答案错误', icon: 'none' })
      }
    },

    selectConceptOption(index) {
      if (this.conceptShowResult) return
      this.conceptSelectedOption = index
      this.conceptShowResult = true
      
      const isCorrect = this.isConceptOptionCorrect(index)
      
      if (isCorrect) {
        uni.showToast({ title: '回答正确！', icon: 'success' })
        this.fireBasicConfetti() // 添加正确时的特效
      } else {
        uni.showToast({ title: '答案错误', icon: 'none' })
      }
    },

    handleNext() {
      if (!this.showConceptQuestion) {
        // 引入题部分
        if (this.isLeadInAnswerCorrect) {
          this.showConceptQuestion = true
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
          // 全部完成时的特效
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
    }
  }
}
</script>

<style>
/* 添加样式以确保特效显示在正确位置 */
.quiz-container {
  position: relative;
  z-index: 1;
}

/* canvas-confetti 会自动创建固定定位的 canvas 元素 */
canvas {
  position: fixed !important;
  z-index: 999 !important;
}
</style> 