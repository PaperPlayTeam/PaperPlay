import request from '@/utils/request'

/**
 * 获取单个问题
 * @param {string} questionId - 问题ID
 * @returns {Promise<{
 *   success: boolean,
 *   message: string,
 *   data: {
 *     id: string,
 *     level_id: string,
 *     stem: string,
 *     content_json: string,
 *     answer_json: string,
 *     score: number,
 *     difficulty: number,
 *     created_by: string,
 *     created_at: string
 *   }
 * }>}
 */
export function getQuestion(questionId) {
  return request({
    url: `/api/v1/questions/${questionId}`,
    method: 'GET'
  })
}

/**
 * 问题类型枚举
 */
export const QuestionType = {
  MCQ: 'mcq',        // 多选题
  ESSAY: 'essay',    // 问答题
  TF: 'true_false'   // 判断题
}

/**
 * 问题难度枚举
 */
export const QuestionDifficulty = {
  EASY: 1,
  MEDIUM: 2,
  HARD: 3,
  EXPERT: 4,
  MASTER: 5
} 