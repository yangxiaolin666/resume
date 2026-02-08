import request from '@/utils/request'

export const getStudentById = (id) => {
  return request.post('/com/detailSearch', {
    id
  })
}

export const completeInfo = ({
  stuId,
  name,
  age,
  gender,
  nation,
  major,
  degree,
  graduationYear,
  internship,
  //languages,
  //languageLevel,
  specialities
}) => {
  return request.put('/stu/completeInfo', {
    stuId,
    name,
    age,
    gender,
    nation,
    major,
    degree,
    graduationYear,
    internship,
    //languages,
    //languageLevel,
    specialities
  })
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
