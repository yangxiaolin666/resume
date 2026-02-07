import request from '@/utils/request'

export const getStudentById = (id) => {
  return request.post('/com/detailSearch', {
    id
  })
}

export const completeInfo = ({
  id,
  stuId,
  name,
  age,
  gender,
  nation,
  major,
  degree,
  graduationYear,
  internship,
  languages,
  languageLevel,
  cities,
  provinces,
  certNames,
  certLevel,
  internshipCompany,
  specialities
}) => {
  return request.put('/stu/completeInfo', {
    id,
    stuId,
    name,
    age,
    gender,
    nation,
    major,
    degree,
    graduationYear,
    internship,
    languages,
    languageLevel,
    cities,
    provinces,
    certNames,
    certLevel,
    internshipCompany,
    specialities
  })
}
