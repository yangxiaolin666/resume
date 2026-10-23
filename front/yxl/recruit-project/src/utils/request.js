import { useUserStore } from '@/stores/user'
import axios from 'axios'
import { ElMessage, ElLoading } from 'element-plus'

// 后端地址：
//   开发环境统一走 Vite 代理 /api-backend（同源，不存在 CORS 问题）；
//   生产环境优先读取 VITE_API_BASE_URL，未设置时使用线上地址。
// 注意：开发环境不再直接请求 http://127.0.0.1:8081，
// 避免浏览器跨域、端口拦截以及 IDE 内嵌预览的混合内容问题。
export const API_BASE_URL = import.meta.env.DEV
  ? '/api-backend'
  : import.meta.env.VITE_API_BASE_URL || 'http://182.92.210.198:8081'

// 后端返回的文件路径可能是 uploads\resume\x.pdf 这种反斜杠形式，统一转成可访问的 URL
export function fileUrl(path) {
  if (!path) return ''
  const text = String(path)
  if (/^https?:\/\//i.test(text)) return text
  return API_BASE_URL + '/' + text.replace(/\\/g, '/').replace(/^\/+/, '')
}

const request = axios.create({
  baseURL: API_BASE_URL,
  timeout: 5000,
  headers: { platform: 'H5' }
})

// 统一把后端的技术性错误转换成用户能看懂的提示
const GENERIC_ERROR = '操作失败，请稍后重试'
const NETWORK_ERROR = '网络连接失败，请检查网络后重试'
const technicalPattern =
  /(json|unmarshal|sql|mysql|driver|dial tcp|deadline|panic|invalid character|unsupported type|interface conversion|reflect|eof|服务器内部错误|数据库)/i

function friendlyMessage(message, fallback = GENERIC_ERROR) {
  const text = String(message || '').trim()
  if (!text) return fallback
  return technicalPattern.test(text) ? fallback : text
}
let loadObj = null
const closeLoading = () => {
  if (loadObj) {
    loadObj.close()
    loadObj = null
  }
}
request.interceptors.request.use(
  function (config) {
    const userStore = useUserStore()
    if (userStore.token) {
      config.headers.Authorization = 'Bearer ' + userStore.token
    }
    if (!loadObj) {
      loadObj = ElLoading.service({
        lock: true,
        text: '加载中...',
        background: 'rgba(0, 0, 0, 0.7)'
      })
    }
    return config
  },
  function (error) {
    closeLoading()
    return Promise.reject(error)
  }
)
request.interceptors.response.use(
  function (res) {
    closeLoading()
    if (res.data.code === 200 || res.data.code === 201) {
      return res
    }
    // config.silent = true 时只返回失败，不弹全局提示（用于“没有简历”这类预期内的空结果）
    if (!res.config?.silent) {
      ElMessage({
        message: friendlyMessage(res.data.message, '服务异常，请稍后重试'),
        type: 'error'
      })
    }
    return Promise.reject(res.data)
  },
  function (error) {
    closeLoading()
    console.error('请求错误详情:', error.response)
    if (!error.config?.silent) {
      ElMessage({
        message: error.response
          ? friendlyMessage(error.response.data?.message, '服务异常，请稍后重试')
          : NETWORK_ERROR,
        type: 'error'
      })
    }

    return Promise.reject(error)
  }
)
export default request
