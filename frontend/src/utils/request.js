import { message } from 'ant-design-vue'

/**
 * 统一处理异步请求的方法
 * @param {Promise} apiCall - 一个返回 Promise 的 API 方法，如 DeviceInfo()
 * @returns {Promise<any | undefined>} 成功时返回 data，失败时返回 undefined，并提示错误
 */
export async function request(apiCall) {
    try {
        const res = await apiCall

        if (res && res.code === 200) {
            return res
        } else {
            const errorMsg = res?.msg || '请求失败'
            message.error(errorMsg)
            return undefined
        }
    } catch (error) {
        console.error('API 请求异常:', error)
        message.error('网络异常，请重试')
        return undefined
    }
}
