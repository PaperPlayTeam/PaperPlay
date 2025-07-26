import request from '@/utils/request'

/**
 * 获取所有学科
 * @returns {Promise<{
 *   success: boolean,
 *   message: string,
 *   data: Array<{
 *     id: string,
 *     name: string,
 *     description: string,
 *     created_at: string,
 *     updated_at: string
 *   }>
 * }>}
 */
export function getAllSubjects() {
  return request({
    url: '/api/v1/subjects',
    method: 'GET'
  })
}

/**
 * 获取学科下的所有论文
 * @param {string} subjectId - 学科ID
 * @returns {Promise<{
 *   success: boolean,
 *   message: string,
 *   data: Array<{
 *     id: string,
 *     subject_id: string,
 *     title: string,
 *     citation: string,
 *     created_at: string,
 *     updated_at: string
 *   }>
 * }>}
 */
export function getSubjectPapers(subjectId) {
  return request({
    url: `/api/v1/subjects/${subjectId}/papers`,
    method: 'GET'
  })
} 