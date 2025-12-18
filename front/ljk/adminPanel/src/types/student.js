/**
 * 学生信息类
 * 用于定义学生数据结构和提供基本操作方法
 */
export class Student {
  constructor({
    id = '',
    stuId = '',
    name = '',
    age = null,
    gender = '',
    nation = '',
    major = '',
    degree = '',
    graduationYear = null,
    internship = null,
    languages = [],
    languageLevel = '',
    cities = [],
    provinces = [],
    certNames = [],
    certLevel = '',
    internshipCompany = '',
    specialities = [],
    education = '',
    internshipCnt = [],
    language = [],
    index = 0,
    ...extra
  } = {}) {
    // 基本信息
    this.id = id; // 唯一标识符
    this.stuId = stuId; // 学生ID
    this.name = name; // 姓名
    this.age = age; // 年龄
    this.gender = gender; // 性别
    this.nation = nation; // 民族
    this.index = index; // 列表索引
    
    // 教育信息
    this.major = major; // 所属专业
    this.degree = degree; // 学位
    this.graduationYear = graduationYear; // 最小毕业年份
    this.education = education; // 学历
    
    // 实习信息
    this.internship = internship; // 实习次数
    this.internshipCnt = internshipCnt; // 实习次数
    this.internshipCompany = internshipCompany; // 实习公司
    
    // 语言能力
    this.languages = languages; // 语言要求
    this.languageLevel = languageLevel; // 语言等级
    this.language = language; // 语言能力
    
    // 地理位置
    this.cities = cities; // 城市
    this.provinces = provinces; // 省份
    
    // 证书信息
    this.certNames = certNames; // 证书名称
    this.certLevel = certLevel; // 证书等级
    
    // 专业技能
    this.specialities = specialities; // 专业技能
    
    // 额外信息
    Object.assign(this, extra);
  }

  /**
   * 获取格式化的性别显示
   */
  getGenderText() {
    return this.gender;
  }

  /**
   * 获取学历信息
   */
  getEducationText() {
    return this.education;
  }

  /**
   * 获取专业信息
   */
  getMajorText() {
    return this.major;
  }

  /**
   * 获取实习次数信息
   */
  getInternshipText() {
    return this.internship;
  }

  /**
   * 获取语言能力信息
   */
  getLanguageText() {
    return this.languages;
  }

  /**
   * 获取语言等级信息
   */
  getLanguageLevelText() {
    return this.languageLevel;
  }

  /**
   * 获取城市信息
   */
  getCitiesText() {
    return this.cities;
  }

  /**
   * 获取省份信息
   */
  getProvincesText() {
    return this.provinces;
  }

  /**
   * 获取证书名称信息
   */
  getCertNamesText() {
    return this.certNames;
  }

  /**
   * 获取证书等级信息
   */
  getCertLevelText() {
    return this.certLevel;
  }

  /**
   * 获取专业技能信息
   */
  getSpecialitiesText() {
    return this.specialities;
  }

  /**
   * 获取学位信息
   */
  getDegreeText() {
    return this.degree;
  }

  /**
   * 获取毕业年份信息
   */
  getGraduationYearText() {
    return this.graduationYear;
  }

  /**
   * 获取实习公司信息
   */
  getInternshipCompanyText() {
    return this.internshipCompany;
  }

  /**
   * 获取民族信息
   */
  getNationText() {
    return this.nation;
  }

  /**
   * 获取年龄信息
   */
  getAgeText() {
    return this.age;
  }

  /**
   * 更新学生信息
   * @param {Object} updates - 要更新的属性
   */
  update(updates) {
    Object.assign(this, updates);
    return this;
  }

  /**
   * 转换为纯对象
   */
  toObject() {
    return {
      id: this.id,
      stuId: this.stuId,
      studentId: this.studentId,
      name: this.name,
      age: this.age,
      gender: this.gender,
      nation: this.nation,
      major: this.major,
      degree: this.degree,
      graduationYear: this.graduationYear,
      internship: this.internship,
      languages: this.languages,
      languageLevel: this.languageLevel,
      cities: this.cities,
      provinces: this.provinces,
      certNames: this.certNames,
      certLevel: this.certLevel,
      internshipCompany: this.internshipCompany,
      specialities: this.specialities,
      education: this.education,
      internshipCnt: this.internshipCnt,
      language: this.language,
      index: this.index
    };
  }

  /**
   * 从对象创建Student实例
   * @param {Object} data - 学生数据对象
   */
  static fromObject(data) {
    return new Student(data);
  }

  /**
   * 批量从对象数组创建Student实例数组
   * @param {Array} dataArray - 学生数据对象数组
   */
  static fromArray(dataArray) {
    return dataArray.map((item, index) => 
      new Student({ ...item, index: index + 1 })
    );
  }
}

/**
 * 学生信息类型定义
 */
export const StudentType = {
  id: String,
  stuId: Number,
  studentId: String,
  name: String,
  age: Number,
  gender: String,
  nation: String,
  major: Array,
  degree: String,
  graduationYear: Number,
  internship: Number,
  languages: Array,
  languageLevel: String,
  cities: Array,
  provinces: Array,
  certNames: Array,
  certLevel: String,
  internshipCompany: String,
  specialities: Array,
  education: Array,
  internshipCnt: Array,
  language: Array,
  index: Number
};
