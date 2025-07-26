import request from '@/utils/request'

/**
 * 获取单篇论文
 * @param {string} paperId - 论文ID
 * @returns {Promise<{
 *   success: boolean,
 *   message: string,
 *   data: {
 *     id: string,
 *     subject_id: string,
 *     title: string,
 *     citation: string,
 *     created_at: string,
 *     updated_at: string
 *   }
 * }>}
 */
export function getPaper(paperId) {
  return request({
    url: `/api/v1/papers/${paperId}`,
    method: 'GET'
  })
}

/**
 * 获取论文对应的关卡
 * @param {string} paperId - 论文ID
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
export function getPaperLevel(paperId) {
  return request({
    url: `/api/v1/papers/${paperId}/level`,
    method: 'GET'
  })
} 