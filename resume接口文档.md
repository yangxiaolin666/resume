# 接口文档

```
StudentQueryDTO {
	"major":"计算机科学与技术",
	"languages":["English", "日语"],
	"pageNum":1,
	"pageSize":10	
}
```



## 企业端("/com")

### 1.查询学生

#### 基本信息

- 请求路径：/search
- 请求方式：POST

#### 请求参数(body)

| 变量名            | 必填 | 数据类型     | 示例               | 备注                                     |
| :---------------- | :--- | :----------- | :----------------- | :--------------------------------------- |
| id                | 否   | Integer      | 12345              | 学生ID                                   |
| stuId             | 否   | Integer      | 20230001           | 学号                                     |
| name              | 否   | String       | "张三"             | 姓名（模糊查询）                         |
| age               | 否   | Integer      | 22                 | 年龄                                     |
| gender            | 否   | String       | "男"               | 性别                                     |
| nation            | 否   | String       | "汉族"             | 民族                                     |
| major             | 否   | String       | "计算机科学"       | 所属专业（模糊查询）                     |
| degree            | 否   | String       | "本科"             | 学位                                     |
| graduationYear    | 否   | Integer      | 2023               | 最小毕业年份                             |
| internship        | 否   | Integer      | 2                  | 实习次数                                 |
| languages         | 否   | List<String> | ["英语", "日语"]   | 语言要求（多值筛选）                     |
| languageLevel     | 否   | String       | "六级"             | 语言等级                                 |
| cities            | 否   | List<String> | ["北京", "上海"]   | 城市（多值筛选）                         |
| provinces         | 否   | List<String> | ["广东", "江苏"]   | 省份（多值筛选）                         |
| certNames         | 否   | List<String> | ["AWS", "PMP"]     | 证书名称（多值筛选）                     |
| certLevel         | 否   | String       | "高级"             | 证书等级                                 |
| internshipCompany | 否   | String       | "腾讯"             | 实习公司                                 |
| specialities      | 否   | List<String> | ["Java", "Python"] | 专业技能（多值筛选）                     |
| pageNum           | 否   | Integer      | 1                  | 当前页码（默认值：1）                    |
| pageSize          | 否   | Integer      | 20                 | 每页大小（默认值：10）                   |
| offset            | 否   | Integer      | 0                  | 偏移量（自动计算：(pageNum-1)*pageSize） |

#### 响应数据

| 字段    | 说明       | 备注          |
| ------- | ---------- | ------------- |
| code    | 状态码     | 200：请求成功 |
| data    | 返回的数据 |               |
| message | 信息       | 成功/错误信息 |

### 2.查询学生详细信息

#### 基本信息

- 请求路径：/detailSearch
- 请求方式：POST

#### 请求参数(body)

| 变量名            | 必填 | 数据类型     | 示例               | 备注                                     |
| :---------------- | :--- | :----------- | :----------------- | :--------------------------------------- |
| id                | 否   | Integer      | 12345              | 学生ID                                   |
| stuId             | 否   | Integer      | 20230001           | 学号                                     |
| name              | 否   | String       | "张三"             | 姓名（模糊查询）                         |
| age               | 否   | Integer      | 22                 | 年龄                                     |
| gender            | 否   | String       | "男"               | 性别                                     |
| nation            | 否   | String       | "汉族"             | 民族                                     |
| major             | 否   | String       | "计算机科学"       | 所属专业（模糊查询）                     |
| degree            | 否   | String       | "本科"             | 学位                                     |
| graduationYear    | 否   | Integer      | 2023               | 最小毕业年份                             |
| internship        | 否   | Integer      | 2                  | 实习次数                                 |
| languages         | 否   | List<String> | ["英语", "日语"]   | 语言要求（多值筛选）                     |
| languageLevel     | 否   | String       | "六级"             | 语言等级                                 |
| cities            | 否   | List<String> | ["北京", "上海"]   | 城市（多值筛选）                         |
| provinces         | 否   | List<String> | ["广东", "江苏"]   | 省份（多值筛选）                         |
| certNames         | 否   | List<String> | ["AWS", "PMP"]     | 证书名称（多值筛选）                     |
| certLevel         | 否   | String       | "高级"             | 证书等级                                 |
| internshipCompany | 否   | String       | "腾讯"             | 实习公司                                 |
| specialities      | 否   | List<String> | ["Java", "Python"] | 专业技能（多值筛选）                     |
| pageNum           | 否   | Integer      | 1                  | 当前页码（默认值：1）                    |
| pageSize          | 否   | Integer      | 20                 | 每页大小（默认值：10）                   |
| offset            | 否   | Integer      | 0                  | 偏移量（自动计算：(pageNum-1)*pageSize） |

#### 响应数据

| 字段    | 说明       | 备注          |
| ------- | ---------- | ------------- |
| code    | 状态码     | 200：请求成功 |
| data    | 返回的数据 |               |
| message | 信息       | 成功/错误信息 |