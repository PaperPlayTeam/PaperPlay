const BASE_URL = 'https://paperplay.zsh.cool';

/**
 * 封装请求方法
 * @param {Object} options - 请求配置
 * @param {string} options.url - 请求地址
 * @param {string} options.method - 请求方法
 * @param {Object} [options.data] - 请求数据
 * @param {Object} [options.header] - 请求头
 * @returns {Promise<any>}
 */
const request = (options = {}) => {
    return new Promise((resolve, reject) => {
        // 处理请求地址
        options.url = BASE_URL + options.url;
        
        // 获取token
        const token = uni.getStorageSync('token');
        
        // 请求头
        options.header = {
            'content-type': 'application/json',
            ...options.header,
            // 如果有token，添加到header
            ...(token ? { 'Authorization': `Bearer ${token}` } : {})
        };
        
        // 发起请求
        uni.request({
            ...options,
            success: (res) => {
                const { statusCode, data } = res;
                
                // 成功状态包括200和201
                if (statusCode === 200 || statusCode === 201) {
                    resolve(data);
                } else if (statusCode === 401) {
                    // token过期或未授权
                    uni.removeStorageSync('token');
                    uni.removeStorageSync('refresh_token');
                    reject({
                        message: '请重新登录'
                    });
                } else {
                    // 其他错误
                    reject({
                        message: data.message || '请求失败'
                    });
                }
            },
            fail: (err) => {
                reject({
                    message: '网络错误',
                    error: err
                });
            }
        });
    });
};

// 导出request方法
export default request;

// 使用示例：
/*
import request from '@/utils/request';

// GET请求
request.get('/api/users', { page: 1 })
    .then(res => {
        console.log(res);
    })
    .catch(err => {
        console.error(err);
    });

// POST请求
request.post('/api/users', {
    name: 'John',
    email: 'john@example.com'
})
    .then(res => {
        console.log(res);
    })
    .catch(err => {
        console.error(err);
    });
*/ 