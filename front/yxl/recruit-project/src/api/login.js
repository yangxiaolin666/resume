import request from '@/utils/request'

export const loginToR = (username, password, userType) => {
  return request.post('/login', {
    username,
    password,
    userType
  })
}
export const changePwd=(old_pwd,new_pwd,re_pwd)=>{
  return request.patch('/stu/updatePwd',{
    "old_pwd":old_pwd,
    "new_pwd":new_pwd,
    "re_pwd":re_pwd
  })
}