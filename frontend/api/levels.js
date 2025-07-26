import request from '@/utils/request';

/**
 * 获取关卡中的问题列表
 * @param {string} levelId - 关卡ID
 * @returns {Promise<{
 *   success: boolean,
 *   message: string,
 *   data: Array<{
 *     id: string,
 *     level_id: string,
 *     stem: string,
 *     content_json: string,
 *     answer_json: string,
 *     score: number,
 *     difficulty: number,
 *     created_by: string,
 *     created_at: string
 *   }>
 * }>}
 */
export function getLevelQuestions(levelId) {
    return request({
        url: `/api/v1/levels/${levelId}/questions`,
        method: 'GET'
    });
}

/**
 * 开始一个关卡
 * @param {string} levelId - 关卡ID
 * @returns {Promise<{
 *   success: boolean,
 *   message: string,
 *   data: {
 *     user_id: string,
 *     level_id: string,
 *     status: number,
 *     started_at: string
 *   }
 * }>}
 */
export function startLevel(levelId) {
    return request({
        url: `/api/v1/levels/${levelId}/start`,
        method: 'POST'
    });
}

/**
 * 提交答案
 * @param {string} levelId - 关卡ID
 * @param {Object} data - 答案数据
 * @param {string} data.question_id - 问题ID
 * @param {Object} data.answer_json - 用户的答案
 * @param {number} data.duration_ms - 答题时长（毫秒）
 * @returns {Promise<{
 *   success: boolean,
 *   message: string,
 *   data: {
 *     question_id: string,
 *     is_correct: boolean,
 *     score: number,
 *     total_score: number
 *   }
 * }>}
 */
export function submitAnswer(levelId, questionId, answerJson, durationMs) {
    return request({
        url: `/api/v1/levels/${levelId}/submit`,
        method: 'POST',
        data: {
            question_id: questionId,
            answer_json: answerJson,
            duration_ms: durationMs
        }
    });
}

/**
 * 完成关卡
 * @param {string} levelId - 关卡ID
 * @returns {Promise<{
 *   success: boolean,
 *   message: string,
 *   data: {
 *     user_id: string,
 *     level_id: string,
 *     score: number,
 *     stars: number,
 *     completed_at: string
 *   }
 * }>}
 */
export function completeLevel(levelId) {
    return request({
        url: `/api/v1/levels/${levelId}/complete`,
        method: 'POST'
    });
}

/**
 * 获取单个关卡信息
 * @param {string} levelId - 关卡ID
 * @returns {Promise<{
 *   success: boolean,
 *   message: string,
 *   data: {
 *     id: string,
 *     paper_id: string,
 *     paper_author: string,
 *     paper_pub_ym: string,
 *     citation_count: number,
 *     name: string,
 *     pass_condition: string,
 *     meta_json: string,
 *     created_at: string,
 *     updated_at: string
 *   }
 * }>}
 */
export function getLevel(levelId) {
    return request({
        url: `/api/v1/levels/${levelId}`,
        method: 'GET'
    });
} 