import request from '@/utils/request'

export const loginToR = (username, password, userType) => {
  return request.post('/login', {
    username,
    password,
    userType
  })
}
