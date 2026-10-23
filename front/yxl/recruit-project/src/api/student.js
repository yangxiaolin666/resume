import request from '@/utils/request'

export const getStudentById = (id) => {
  return request.post('/com/detailSearch', {
    id
  })
}
export const getStudentDetailById = (id) => {
  return request.post('/com/detailSearch', {
    id
  })
}
export const getAllStudent=(data)=>{
  return request.post('/com/detailSearch',
    data
  )
}
export const getStatics=()=>{
  return request.get('/stu/statistics',{})
}
const completeObj={
  "age": 40,
  "gender": "男",
  "nation": "汉族",
  "major": "计算机科学与技术",
  "degree": "本科生",
  "graduationYear": 2027,
  "internship": 1,
  "languages": [
    {
      "language": "英语",
      "level": "四级"
    }
  ],
  "cities": [
    "大连","北京"
  ],
  "provinces": [
    "辽宁","广东"
  ],
  "certificates": [
    {
      "certName": "计算机",
      "level": "二级"
    }
  ],
  "internships": [
    {
      "company": "腾讯",
      "experience": "游戏后端开发"
    },
    {
      "company": "阿里",
      "experience": "游戏后端开发"
    },
  ],
  "specialities": [
    "Java","C++","Python"
  ],
  "jobs": [
    "后端开发","前端开发"
  ]
}
export const completeInfo = (data) => {
  return request.put('/stu/completeInfo',data)
}
export const updateResume=((resumeFile)=>{
  const formData = new FormData()
  formData.append('resumeFile', resumeFile)
  return request.post('/stu/uploadResume',
    formData
  )
})

export const getResume=(()=>{
  return request.get('/uploads/resume')
})

// 当前学生最新上传的简历（聊天里「发送简历」使用）
// config 可传 { silent: true }，页面初始化时没简历也不弹错误提示
export const getMyResume = (config = {}) => request.get('/stu/resume/latest', config)
export const addLike=(list)=>{
  return request.post('/com/LikeAdd',
    list
  )
}
export const deleteLike=(list)=>{
  return request.delete('/com/LikeDelete',{
      data:list
    }
  )
}
export const queryLike=(data)=>{
  return request.post('/com/LikeQuery',data)
}
export const queryComId=()=>{
  return request.get('/com/QueryComIdByUserId')
}