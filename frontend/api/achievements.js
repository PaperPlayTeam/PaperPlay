import request from '@/utils/request'

/**
 * 获取所有成就
 * @returns {Promise<{
 *   success: boolean,
 *   data: Array<{
 *     id: string,
 *     name: string,
 *     description: string,
 *     icon_url: string,
 *     level: number,
 *     category: string,
 *     is_active: boolean,
 *     rules: {
 *       type: string,
 *       conditions: Array<{
 *         field: string,
 *         operator: string,
 *         value: any
 *       }>
 *     },
 *     nft_metadata: {
 *       name: string,
 *       description: string,
 *       image: string,
 *       attributes: Array<{
 *         trait_type: string,
 *         value: string
 *       }>
 *     }
 *   }>
 * }>}
 */
export function getAllAchievements() {
  return request({
    url: '/api/v1/achievements',
    method: 'GET'
  })
}

/**
 * 获取用户已获得的成就
 * @returns {Promise<{
 *   success: boolean,
 *   data: Array<{
 *     id: string,
 *     user_id: string,
 *     achievement_id: string,
 *     earned_at: string,
 *     event_data: {
 *       trigger_event: string,
 *       level_id?: string
 *     },
 *     achievement: {
 *       id: string,
 *       name: string,
 *       description: string,
 *       icon_url: string,
 *       level: number,
 *       is_active: boolean,
 *       nft_metadata: {
 *         name: string,
 *         description: string,
 *         image: string
 *       }
 *     }
 *   }>
 * }>}
 */
export function getUserAchievements() {
  return request({
    url: '/api/v1/users/achievements',
    method: 'GET'
  })
}

/**
 * 触发成就评估
 * @returns {Promise<{
 *   success: boolean,
 *   message: string
 * }>}
 */
export function evaluateAchievements() {
  return request({
    url: '/api/v1/achievements/evaluate',
    method: 'POST'
  })
}

/**
 * 成就类型枚举
 */
export const AchievementCategory = {
  PROGRESSION: 'progression',  // 进度类成就
  SKILL: 'skill',             // 技能类成就
  COLLECTION: 'collection',    // 收集类成就
  SOCIAL: 'social',           // 社交类成就
  CHALLENGE: 'challenge'      // 挑战类成就
}

/**
 * 成就规则类型枚举
 */
export const AchievementRuleType = {
  FIRST_TRY: 'first_try',     // 首次尝试
  COUNTER: 'counter',         // 计数器
  COLLECTION: 'collection',    // 收集
  TIME_BASED: 'time_based',   // 基于时间
  COMBO: 'combo'              // 连击
}

/**
 * 成就等级枚举
 */
export const AchievementLevel = {
  BRONZE: 1,   // 铜牌
  SILVER: 2,   // 银牌
  GOLD: 3      // 金牌
} 