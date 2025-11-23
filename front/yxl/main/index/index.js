// 创建Vue应用实例
const app = Vue.createApp({
    data() {
        return {
            students: [
                {
                    id: 1,
                    name: "杨潇林",
                    age: 19,
                    education: "本科",
                    major: "计算机科学与技术",
                    school: "大连民族大学",
                    experience: "应届毕业生",
                    skills: ["JavaScript", "Vue", "HTML/CSS"],
                    languages: ["英语四级"],
                    certificates: ["软件设计师"],
                    internships: 0,
                    computerSkills: ["Java"],
                    jobRegions: ["辽宁省大连市"],
                    jobPositions: ["软件开发"],
                    bio: "热爱前端开发，有多个项目经验，熟悉现代前端框架。",
                    contactInfo: { phone: "13800138001", email: "yangxl@example.com" },
                    photo: ""
                },
                {
                    id: 2,
                    name: "刘世鹏",
                    age: 19,
                    education: "研究生",
                    major: "软件工程",
                    school: "大连民族大学",
                    experience: "1年经验",
                    skills: ["React", "TypeScript", "Node.js"],
                    languages: ["英语四级", "日语N2"],
                    certificates: ["程序员"],
                    internships: 1,
                    computerSkills: ["Java"],
                    jobRegions: ["辽宁省大连市"],
                    jobPositions: ["软件开发", "软件技术服务"],
                    bio: "有1年全栈开发经验，擅长React生态系统，对后端也有一定了解。",
                    contactInfo: { phone: "13800138002", email: "liusp@example.com" },
                    photo: ""
                },
                {
                    id: 3,
                    name: "范茂阳",
                    age: 19,
                    education: "本科",
                    major: "软件工程",
                    school: "大连民族大学",
                    experience: "应届毕业生",
                    skills: ["Python", "C++", "机器学习"],
                    languages: ["英语四级"],
                    certificates: [],
                    internships: 0,
                    computerSkills: [],
                    jobRegions: [],
                    jobPositions: [],
                    bio: "对人工智能和机器学习有浓厚兴趣，参与过相关项目开发。",
                    contactInfo: { phone: "13800138003", email: "fanmy@example.com" },
                    photo: ""
                },
                {
                    id: 4,
                    name: "人机",
                    age: 19,
                    education: "本科",
                    major: "人工智能",
                    school: "大连民族大学",
                    experience: "2年经验",
                    skills: ["Java", "Spring Boot", "MySQL"],
                    languages: ["英语四级", "日语N1"],
                    certificates: ["网络管理员"],
                    internships: 1,
                    computerSkills: ["Java"],
                    jobRegions: ["辽宁省大连市"],
                    jobPositions: ["软件开发", "网络安全维护"],
                    bio: "后端开发工程师，有扎实的Java基础，熟悉主流框架和数据库。",
                    contactInfo: { phone: "13800138004", email: "renji@example.com" },
                    photo: ""
                },
                {
                    id: 5,
                    name: "网络",
                    age: 19,
                    education: "本科",
                    major: "网络工程",
                    school: "大连民族大学",
                    experience: "应届毕业生",
                    skills: ["UI设计", "Photoshop", "Figma"],
                    languages: ["英语四级", "韩语TOPIK I"],
                    certificates: [],
                    internships: 0,
                    computerSkills: [],
                    jobRegions: [],
                    jobPositions: ["软件技术服务"],
                    bio: "UI设计师，有良好的美学素养和用户体验意识，熟练使用设计工具。",
                    contactInfo: { phone: "13800138005", email: "wangluo@example.com" },
                    photo: ""
                },
                
            ],
            
            // 筛选选项
            educationOptions: ["本科","研究生"],
            skillOptions: ["JavaScript", "Vue", "React", "TypeScript", "Node.js", "Python", "C++", "Java", "Spring Boot", "MySQL", "HTML/CSS", "UI设计", "Photoshop", "Figma", "机器学习"],
            experienceOptions: ["应届毕业生", "1年经验", "2年经验", "3年以上"],
            languageOptions: ["英语四级", "日语N1", "日语N2", "韩语TOPIK I", "韩语TOPIK II"],
            certificateOptions: ["软件设计师", "程序员", "网络管理员"],
            internshipOptions: ["0次", "1次", "2次及以上"],
            computerSkillOptions: ["Java"],
            jobRegionOptions: ["辽宁省大连市"],
            jobPositionOptions: ["网络安全维护", "软件开发", "软件技术服务"],
            
            // 自定义标签相关
            customLanguage: '',
            customCertificate: '',
            customComputerSkill: '',
            customJobRegion: '',
            customJobPosition: '',
            
            // 选中的筛选条件
            selectedFilters: {
                education: '',
                majors: [],
                skills: [],
                experience: [],
                languages: [],
                certificates: [],
                internships: [],
                computerSkills: [],
                jobRegions: [],
                jobPositions: []
            },
            
            // 原始专业选项
            originalMajorOptions: ["计算机科学与技术", "软件工程", "网络工程", "人工智能"],
            // 研究生专业选项
            postgraduateMajorOptions: ["计算机科学与技术", "人工智能"],
            
            // 当前查看详情的学生
            currentStudent: null,
            // 动画版本号，用于强制重新渲染动画
            animationVersion: 0,
            // 用于触发动画的key
            animationKey: 0,
        }
    },
    
    computed: {
            // 根据筛选条件过滤学生并添加动画索引
            filteredStudents() {
                return this.students.filter(student => {
                    // 学历筛选
                    if (this.selectedFilters.education && 
                        this.selectedFilters.education !== student.education) {
                        return false;
                    }
                    
                    // 专业筛选
                    if (this.selectedFilters.majors.length > 0 && 
                        !this.selectedFilters.majors.includes(student.major)) {
                        return false;
                    }
                    
                    // 技能筛选
                    if (this.selectedFilters.skills.length > 0) {
                        const hasSkill = this.selectedFilters.skills.some(skill => 
                            student.skills.includes(skill)
                        );
                        if (!hasSkill) {
                            return false;
                        }
                    }
                    
                    // 工作经验筛选
                    if (this.selectedFilters.experience.length > 0 && 
                        !this.selectedFilters.experience.includes(student.experience)) {
                        return false;
                    }
                    
                    // 语言掌握情况筛选
                    if (this.selectedFilters.languages.length > 0) {
                        const hasLanguage = this.selectedFilters.languages.some(lang => 
                            student.languages.includes(lang)
                        );
                        if (!hasLanguage) {
                            return false;
                        }
                    }
                    
                    // 证书筛选
                    if (this.selectedFilters.certificates.length > 0) {
                        const hasCertificate = this.selectedFilters.certificates.some(cert => 
                            student.certificates.includes(cert)
                        );
                        if (!hasCertificate) {
                            return false;
                        }
                    }
                    
                    // 实习次数筛选
                    if (this.selectedFilters.internships.length > 0) {
                        let matchesInternship = false;
                        if (this.selectedFilters.internships.includes('0次') && student.internships === 0) {
                            matchesInternship = true;
                        }
                        if (this.selectedFilters.internships.includes('1次') && student.internships === 1) {
                            matchesInternship = true;
                        }
                        if (this.selectedFilters.internships.includes('2次及以上') && student.internships >= 2) {
                            matchesInternship = true;
                        }
                        if (!matchesInternship) {
                            return false;
                        }
                    }
                    
                    // 计算机专业技能筛选
                    if (this.selectedFilters.computerSkills.length > 0) {
                        const hasComputerSkill = this.selectedFilters.computerSkills.some(skill => 
                            student.computerSkills.includes(skill)
                        );
                        if (!hasComputerSkill) {
                            return false;
                        }
                    }
                    
                    // 就业意向地区筛选
                    if (this.selectedFilters.jobRegions.length > 0) {
                        const hasJobRegion = this.selectedFilters.jobRegions.some(region => 
                            student.jobRegions.includes(region)
                        );
                        if (!hasJobRegion) {
                            return false;
                        }
                    }
                    
                    // 就业岗位筛选
                    if (this.selectedFilters.jobPositions.length > 0) {
                        const hasJobPosition = this.selectedFilters.jobPositions.some(position => 
                            student.jobPositions.includes(position)
                        );
                        if (!hasJobPosition) {
                            return false;
                        }
                    }
                    
                    return true;
                });
            },
            
            // 根据选择的学历动态决定显示的专业选项
            majorOptions() {
                if (this.selectedFilters.education === '研究生') {
                    return this.postgraduateMajorOptions;
                }
                return this.originalMajorOptions;
            },
            
            // 计算已选择的筛选条件数量
            appliedFilters() {
                let count = 0;
                Object.values(this.selectedFilters).forEach(filter => {
                    if (Array.isArray(filter)) {
                        count += filter.length;
                    } else if (filter) {
                        count++;
                    }
                });
                return count;
            }
    },
    
    methods: {
        resetFilters() {
            this.selectedFilters = {
                education: '',
                majors: [],
                skills: [],
                experience: [],
                languages: [],
                certificates: [],
                internships: [],
                computerSkills: [],
                jobRegions: [],
                jobPositions: []
            };
        },
        
        // 处理学历选择变化，清空已选专业
            handleEducationChange() {
                this.selectedFilters.majors = [];
                // 更新动画版本号，强制重新渲染动画
                this.triggerAnimation();
            },
            
            // 触发动画重新渲染
            triggerAnimation() {
                // 先移除所有行的动画类
                this.animationKey++;
            },
        
       
        // 重置新学生表单
        resetNewStudentForm() {
            this.newStudent = {
                name: '',
                age: '',
                education: '',
                major: '',
                school: '',
                experience: '',
                skills: [],
                languages: [],
                certificates: [],
                internships: 0,
                computerSkills: [],
                jobRegions: [],
                jobPositions: [],
                bio: '',
                contactInfo: {
                    phone: '',
                    email: ''
                },
                photo: null
            };
              
            // 清空自定义标签输入框
            this.customLanguage = '';
            this.customCertificate = '';
            this.customComputerSkill = '';
            this.customJobRegion = '';
            this.customJobPosition = '';
            
            // 清空文件输入框
            const fileInput = document.getElementById('photo-upload');
            if (fileInput) {
                fileInput.value = '';
            }
        },
        
        
    },
});
// 挂载Vue应用
app.mount('#app');
