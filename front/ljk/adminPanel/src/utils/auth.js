// src/utils/auth.js

export const setToken = (token) => {
  localStorage.setItem('token', token)
}


export const getToken = () => {
  return localStorage.getItem('token') || null
}


export const removeToken = () => {
  localStorage.removeItem('token')
}

/**
 * 校验 Token 是否有效（核心逻辑，需和后端/登录同事确认校验规则）
 * @param {string} token - 待校验的 Token
 * @returns {Promise<boolean>} - 是否有效（异步返回，模拟接口校验）
 */
export const validateToken = async (token) => {
  if (!token) return false

  try {
    // ******** 核心：替换为真实的 Token 校验接口 ********
    // 请和后端/登录同事确认：Token 校验的接口地址、请求方式、参数
    const response = await fetch('/api/validate-token', {
      method: 'POST', // 或 GET，按后端要求
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}` // 常见的 Token 传递方式
      },
      body: JSON.stringify({ token }) // 部分后端需要传 token 到 body
    })

    const res = await response.json()
    // 假设后端返回 { code: 200, success: true } 表示 Token 有效
    return res.success && res.code === 200
  } catch (error) {
    console.error('Token 校验失败：', error)
    return false
  }

  // 【本地临时模拟】如果暂时没有接口，可先模拟校验（测试用，上线前替换）
  // return new Promise((resolve) => {
  //   // 模拟 Token 有效（比如判断 Token 非空且长度 > 10）
  //   resolve(token && token.length > 10)
  // })
}