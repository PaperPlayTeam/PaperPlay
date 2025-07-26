import request from '@/utils/request'

/**
 * 获取用户资料
 * @returns {Promise}
 */
export function getUserProfile() {
  return request({
    url: '/api/v1/users/profile',
    method: 'GET'
  })
}

/**
 * 更新用户资料
 * @param {Object} data - 更新的用户数据
 * @param {string} [data.nickname] - 昵称
 * @param {string} [data.avatar] - 头像
 * @param {string} [data.email] - 邮箱
 * @returns {Promise}
 */
export function updateUserProfile(data) {
  return request({
    url: '/api/v1/users/profile',
    method: 'PUT',
    data
  })
}

/**
 * 获取用户学习进度
 * @returns {Promise}
 */
export function getUserProgress() {
  return request({
    url: '/api/v1/users/progress',
    method: 'GET'
  })
}

/**
 * 获取用户成就列表
 * @returns {Promise}
 */
export function getUserAchievements() {
  return request({
    url: '/api/v1/users/achievements',
    method: 'GET'
  })
} 