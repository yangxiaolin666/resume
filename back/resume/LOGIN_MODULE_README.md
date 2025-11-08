# 登录模块说明文档

## 模块概述

本登录模块实现了基于JWT的用户身份认证功能，支持学生、企业、管理员三种用户类型的登录。

## 目录结构

```
src/main/java/com/resume/
├── model/
│   ├── entity/
│   │   └── User.java                    # 用户实体类
│   └── dto/
│       ├── LoginDTO.java                # 登录请求DTO
│       └── LoginVO.java                 # 登录响应VO
├── user/
│   ├── controller/
│   │   └── LoginController.java         # 登录控制器
│   ├── service/
│   │   ├── UserService.java             # 用户服务接口
│   │   └── impl/
│   │       └── UserServiceImpl.java     # 用户服务实现
│   └── mapper/
│       └── UserMapper.java              # 用户数据访问接口
└── utils/
    └── JwtUtil.java                     # JWT工具类
```

## 数据库设计

### user表结构

| 字段名      | 类型         | 说明                         |
| ----------- | ------------ | ---------------------------- |
| user_id     | INT          | 用户ID（主键，自增）         |
| username    | VARCHAR(100) | 用户名（唯一，不能为空）     |
| password    | VARCHAR(50)  | 密码（不能为空）             |
| user_type   | VARCHAR(25)  | 用户类型（student/company/admin） |
| create_time | DATETIME     | 创建时间                     |
| update_time | DATETIME     | 更新时间                     |

## API接口

### 登录接口

**请求地址：** `POST /login`

**请求体：**
```json
{
  "username": "20230001",
  "password": "20230001"
}
```

**成功响应：**
```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "userId": 1,
    "username": "20230001",
    "userType": "student",
    "token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9..."
  }
}
```

**错误响应：**

1. 用户名或密码错误（40005）：
```json
{
  "code": 40005,
  "message": "用户名或密码错误",
  "data": null
}
```

2. 登录尝试次数过多（42901）：
```json
{
  "code": 42901,
  "message": "登录尝试次数过多，请30分钟后再试",
  "data": null
}
```

## 功能特性

### 1. 安全特性

- **密码验证**：验证用户名和密码是否匹配
- **登录限制**：5次失败登录后锁定30分钟
- **JWT认证**：使用JWT生成安全token，有效期24小时

### 2. Redis缓存

使用Redis记录登录尝试次数，防止暴力破解：
- Key格式：`login_attempt:{username}`
- 锁定时间：30分钟
- 最大尝试次数：5次

### 3. JWT Token

Token包含以下信息：
- userId：用户ID
- username：用户名
- userType：用户类型
- exp：过期时间（24小时）

## 测试数据

在 `test_user_data.sql` 文件中提供了测试用户数据：

| 用户名     | 密码        | 用户类型 |
| ---------- | ----------- | -------- |
| 20230001   | 20230001    | student  |
| company01  | company01   | company  |
| admin      | admin123    | admin    |

## 使用方法

### 1. 初始化数据库

执行以下SQL文件：
```bash
# 创建数据库和表
mysql -u root -p < resume.sql

# 插入测试数据
mysql -u root -p < test_user_data.sql
```

### 2. 配置Redis

确保Redis服务正在运行，并在 `application.yml` 中配置：
```yaml
spring:
  data:
    redis:
      host: localhost
      port: 6379
```

### 3. 启动应用

```bash
mvn spring-boot:run
```

### 4. 测试登录

使用Postman或curl测试：

```bash
curl -X POST http://localhost:8081/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "20230001",
    "password": "20230001"
  }'
```

## 后续使用Token

在需要认证的接口中，在请求头添加：
```
Authorization: Bearer {token}
```

示例：
```bash
curl -X POST http://localhost:8081/com/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9..." \
  -d '{...}'
```

## 扩展说明

如果需要添加账号禁用功能，可以：

1. 在user表中添加 `is_disabled` 字段：
```sql
ALTER TABLE user ADD COLUMN is_disabled TINYINT(1) DEFAULT 0 COMMENT '是否禁用 0-正常 1-禁用';
```

2. 在 `UserServiceImpl.login()` 方法中添加检查：
```java
// 检查账号状态
if (user.getIsDisabled() != null && user.getIsDisabled() == 1) {
    throw new RuntimeException("40006:账号已被禁用");
}
```

## 注意事项

1. 密码目前使用明文存储，生产环境建议使用BCrypt等加密方式
2. JWT密钥建议配置在配置文件中，不要硬编码
3. Token有效期根据实际需求调整
4. Redis连接失败时需要添加降级处理
