import request from '@/utils/request'

// 发布招聘信息
export const createJobPosting = (data) => {
  return request.post('/com/jobPosting', data)
}

// 修改招聘信息
export const updateJobPosting = (id, data) => {
  return request.put(`/com/jobPosting/${id}`, data)
}

// 删除招聘信息
export const deleteJobPosting = (id) => {
  return request.delete(`/com/jobPosting/${id}`)
}

// 查询招聘信息列表（分页）
export const getJobPostings = (params) => {
  return request.get('/com/jobPostings', { params })
}

// 查询企业所有招聘信息（无分页）
export const getMyJobPostings = (params) => {
  return request.get('/com/myJobPostings', { params })
}

// 查询招聘信息详情
export const getJobPostingDetail = (id) => {
  return request.get(`/com/jobPosting/${id}`)
}

// 启用/禁用招聘信息
export const updateJobPostingStatus = (id, status) => {
  return request.patch(`/com/jobPosting/${id}/status`, { status })
}