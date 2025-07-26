import request from '@/utils/request'

/**
 * 用户登录
 * @param {Object} data - 登录数据
 * @param {string} data.email - 用户邮箱
 * @param {string} data.password - 密码
 * @returns {Promise<{
 *   user: {
 *     id: string,
 *     email: string,
 *     display_name: string,
 *     avatar_url: string,
 *     eth_address: string,
 *     created_at: string,
 *     updated_at: string
 *   },
 *   access_token: string,
 *   refresh_token: string,
 *   token_type: string,
 *   expires_in: number
 * }>}
 */
export function login(data) {
  return request({
    url: '/api/v1/auth/login',
    method: 'POST',
    data
  })
}

/**
 * 用户注册
 * @param {Object} data - 注册数据
 * @param {string} data.email - 用户邮箱
 * @param {string} data.password - 密码
 * @param {string} data.display_name - 显示名称
 * @returns {Promise<{
 *   user: {
 *     id: string,
 *     email: string,
 *     display_name: string,
 *     avatar_url: string,
 *     eth_address: string,
 *     created_at: string,
 *     updated_at: string
 *   },
 *   access_token: string,
 *   refresh_token: string,
 *   token_type: string,
 *   expires_in: number
 * }>}
 */
export function register(data) {
  return request({
    url: '/api/v1/auth/register',
    method: 'POST',
    data
  })
}

/**
 * 刷新访问令牌
 * @returns {Promise<{
 *   access_token: string,
 *   refresh_token: string,
 *   token_type: string,
 *   expires_in: number
 * }>}
 */
export function refreshToken() {
  return request({
    url: '/api/v1/auth/refresh',
    method: 'POST'
  })
}

/**
 * 用户登出
 * @returns {Promise<void>}
 */
export function logout() {
  return request({
    url: '/api/v1/users/logout',
    method: 'POST'
  })
} 