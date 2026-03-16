# Token 验证接口说明

## 接口概述

在用户模块（`/user`）下实现了用于验证 Token 的接口。

## 接口列表

### POST /user/validateToken - 验证 Token 是否有效且未过期

**接口地址：** `POST /user/validateToken`

**功能说明：** 
- 验证 Token 的有效性
- 检查 Token 是否即将过期（30 分钟内）
- 返回 Token 中的用户信息

**请求参数：**

支持两种方式传递 Token：
1. 从请求体传递（推荐）
2. 从请求头传递（需要在拦截器中放行此接口）

**请求格式 1 - 从请求体传递 Token：**

```json
{
  "token": "your_jwt_token_here"
}
```

**请求格式 2 - 不传参数（需要从请求头获取 Token）：**

```json
{}
```

**响应数据：**

| 字段 | 类型 | 说明 |
|------|------|------|
| valid | Boolean | Token 是否有效 |
| message | String | 验证结果消息 |
| userId | Integer | 用户 ID（仅当 valid=true 时返回） |
| userType | String | 用户类型（仅当 valid=true 时返回） |
| needRefresh | Boolean | 是否需要刷新（剩余时间少于 30 分钟） |

**成功响应示例：**

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "valid": true,
    "message": "Token 有效且未过期",
    "userId": 123,
    "userType": "student",
    "needRefresh": false
  }
}
```

**失败响应示例：**

```json
{
  "code": 401,
  "message": "Token 不能为空",
  "data": null
}
```

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "valid": false,
    "message": "Token 无效或已过期",
    "userId": null,
    "userType": null,
    "needRefresh": null
  }
}
```

---

## 使用场景

### 场景 1：前端定期检查 Token 状态

``javascript
// 每隔一段时间检查 Token 是否即将过期
async function checkTokenStatus() {
  const response = await fetch('/user/validateToken', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      token: localStorage.getItem('token')
    })
  });
  
  const result = await response.json();
  
  if (result.data.valid) {
    if (result.data.needRefresh) {
      // Token 快过期了，需要刷新
      await refreshToken();
    }
  } else {
    // Token 已失效，需要重新登录
    redirectToLogin();
  }
}
```

### 场景 2：进入系统前验证 Token

``javascript
// 在进入需要权限的页面之前验证 Token
async function verifyTokenBeforeEnter() {
  const response = await fetch('/user/validateToken', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      token: localStorage.getItem('token')
    })
  });
  
  const result = await response.json();
  
  if (!result.data.valid) {
    // Token 无效，跳转到登录页
    window.location.href = '/login';
  }
}
```

### 场景 3：获取当前用户信息

``javascript
// 获取当前登录用户的信息
async function getCurrentUserInfo() {
  const response = await fetch('/common/getCurrentUser', {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${localStorage.getItem('token')}`
    }
  });
  
  const result = await response.json();
  
  if (result.code === 200) {
    // 可以在这里根据 userType 做不同的页面展示
    console.log('当前用户:', result.data);
  }
}
```

---

## 注意事项

1. **Token 传递方式**：
   - 推荐使用请求体传递 Token
   - 如果使用请求头传递，需要确保该接口在拦截器中被放行

2. **Token 有效期**：
   - Access Token：2 小时
   - Refresh Token：7 天
   - 建议在 Token 到期前 30 分钟进行刷新

3. **安全性**：
   - 所有接口都需要 HTTPS 传输
   - Token 应该存储在安全的地方（如 httpOnly cookie 或加密的 localStorage）

4. **错误处理**：
   - 401：Token 无效或已过期
   - 500：服务器内部错误
   - 其他错误码：查看具体错误消息

---

## 测试示例

### 使用 Postman 测试

**测试 validateToken 接口：**

1. 请求方法：POST
2. 请求 URL：http://localhost:8080/common/validateToken
3. Headers：Content-Type: application/json
4. Body（raw JSON）：
```json
{
  "token": "your_test_token"
}
```

**测试 getCurrentUser 接口：**

1. 请求方法：GET
2. 请求 URL：http://localhost:8080/common/getCurrentUser
3. Headers：Authorization: Bearer your_test_token

### 使用 cURL 测试

```
# 测试 validateToken 接口
curl -X POST http://localhost:8080/user/validateToken \
  -H "Content-Type: application/json" \
  -d '{"token":"your_test_token"}'

# 测试 getCurrentUser 接口
curl -X GET http://localhost:8080/common/getCurrentUser \
  -H "Authorization: Bearer your_test_token"
```
