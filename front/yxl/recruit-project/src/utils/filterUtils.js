// 筛选逻辑封装工具

/**
 * 根据筛选条件过滤学生数据
 * @param {Array} students - 原始学生数组
 * @param {Object} filters - 筛选条件
 * @returns {Array} 过滤后的学生数组
 */
export function filterStudents(students, filters) {
  return students.filter(student => {
    // 姓名模糊查询
    if (filters.name && !student.name.includes(filters.name)) {
      return false
    }

    // 专业模糊查询
    if (filters.major && !student.major.includes(filters.major)) {
      return false
    }

    // 学位筛选
    if (filters.degree && student.degree !== filters.degree) {
      return false
    }

    // 毕业年份筛选
    if (filters.graduationYear && student.graduationYear < filters.graduationYear) {
      return false
    }

    // 实习次数筛选
    if (filters.internship !== null && student.internship < filters.internship) {
      return false
    }

    // 语言多值筛选
    if (filters.languages && filters.languages.length > 0) {
      const hasLanguage = filters.languages.some(lang => 
        student.languages.includes(lang)
      )
      if (!hasLanguage) {
        return false
      }
    }

    // 语言等级筛选
    if (filters.languageLevel && student.languageLevel !== filters.languageLevel) {
      return false
    }

    // 城市多值筛选
    if (filters.cities && filters.cities.length > 0) {
      const hasCity = filters.cities.some(city => 
        student.cities.includes(city)
      )
      if (!hasCity) {
        return false
      }
    }

    // 省份多值筛选
    if (filters.provinces && filters.provinces.length > 0) {
      const hasProvince = filters.provinces.some(province => 
        student.provinces.includes(province)
      )
      if (!hasProvince) {
        return false
      }
    }

    // 证书名称多值筛选
    if (filters.certNames && filters.certNames.length > 0) {
      const hasCert = filters.certNames.some(cert => 
        student.certNames.includes(cert)
      )
      if (!hasCert) {
        return false
      }
    }

    // 证书等级筛选
    if (filters.certLevel && student.certLevel !== filters.certLevel) {
      return false
    }

    // 实习公司筛选
    if (filters.internshipCompany && !student.internshipCompany.includes(filters.internshipCompany)) {
      return false
    }

    // 专业技能多值筛选
    if (filters.specialities && filters.specialities.length > 0) {
      const hasSkill = filters.specialities.some(skill => 
        student.specialities.includes(skill)
      )
      if (!hasSkill) {
        return false
      }
    }

    return true
  })
}

/**
 * 分页处理
 * @param {Array} students - 学生数组
 * @param {number} pageNum - 当前页码
 * @param {number} pageSize - 每页大小
 * @returns {Array} 分页后的学生数组
 */
export function paginateStudents(students, pageNum = 1, pageSize = 10) {
  const startIndex = (pageNum - 1) * pageSize
  const endIndex = startIndex + pageSize
  return students.slice(startIndex, endIndex)
}

/**
 * 组合筛选和分页
 * @param {Array} students - 原始学生数组
 * @param {Object} filters - 筛选条件
 * @param {number} pageNum - 当前页码
 * @param {number} pageSize - 每页大小
 * @returns {Object} 包含分页数据和总数的对象
 */
export function getFilteredAndPaginatedStudents(students, filters, pageNum = 1, pageSize = 10) {
  const filteredStudents = filterStudents(students, filters)
  const paginatedStudents = paginateStudents(filteredStudents, pageNum, pageSize)
  
  return {
    data: paginatedStudents,
    total: filteredStudents.length,
    pageNum,
    pageSize,
    totalPages: Math.ceil(filteredStudents.length / pageSize)
  }
}

/**
 * 从学生数据中提取唯一值
 * @param {Array} students - 学生数组
 * @param {string} field - 字段名
 * @returns {Array} 唯一值数组
 */
export function getUniqueValues(students, field) {
  const values = students.flatMap(student => {
    const value = student[field]
    return Array.isArray(value) ? value : (value ? [value] : [])
  })
  return [...new Set(values)].filter(Boolean)
}

/**
 * 生成筛选选项
 * @param {Array} students - 学生数组
 * @returns {Object} 包含所有筛选选项的对象
 */
export function generateFilterOptions(students) {
  return {
    majors: getUniqueValues(students, 'major'),
    degrees: getUniqueValues(students, 'degree'),
    languages: getUniqueValues(students, 'languages'),
    languageLevels: getUniqueValues(students, 'languageLevel'),
    cities: getUniqueValues(students, 'cities'),
    provinces: getUniqueValues(students, 'provinces'),
    certNames: getUniqueValues(students, 'certNames'),
    certLevels: getUniqueValues(students, 'certLevel'),
    specialities: getUniqueValues(students, 'specialities')
  }
}

/**
 * 重置筛选条件
 * @returns {Object} 默认筛选条件
 */
export function getDefaultFilters() {
  return {
    name: '',
    major: '',
    degree: '',
    graduationYear: null,
    internship: null,
    languages: [],
    languageLevel: '',
    cities: [],
    provinces: [],
    certNames: [],
    certLevel: '',
    internshipCompany: '',
    specialities: []
  }
}

/**
 * 验证筛选条件
 * @param {Object} filters - 筛选条件
 * @returns {Object} 验证结果
 */
export function validateFilters(filters) {
  const errors = {}
  
  if (filters.graduationYear && isNaN(filters.graduationYear)) {
    errors.graduationYear = '毕业年份必须是数字'
  }
  
  if (filters.internship !== null && isNaN(filters.internship)) {
    errors.internship = '实习次数必须是数字'
  }
  
  return {
    valid: Object.keys(errors).length === 0,
    errors
  }
}
