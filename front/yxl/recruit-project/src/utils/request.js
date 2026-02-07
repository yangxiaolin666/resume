import { useUserStore } from '@/stores/user'
import axios from 'axios'
import { ElMessage, ElLoading } from 'element-plus'
const request = axios.create({
  baseURL: 'http://81.70.199.174:8081',
  timeout: 5000,
  headers: { platform: 'H5' }
})
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
    if (res.data.code === 200) {
      return res
    }

    ElMessage({
      message: res.data.message || '服务异常',
      type: 'error'
    })
    return Promise.reject(res.data)
  },
  function (error) {
    closeLoading()
    console.error('请求错误详情:', error.response)
    ElMessage({
      message: error.response?.data?.message || '服务异常',
      type: 'error'
    })

    return Promise.reject(error)
  }
)
export default request
