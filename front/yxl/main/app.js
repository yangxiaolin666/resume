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
                    bio: "热爱前端开发，有多个项目经验，熟悉现代前端框架。"
                },
                {
                    id: 2,
                    name: "刘世鹏",
                    age: 19,
                    education: "硕士",
                    major: "软件工程",
                    school: "大连民族大学",
                    experience: "1年经验",
                    skills: ["React", "TypeScript", "Node.js"],
                    bio: "有1年全栈开发经验，擅长React生态系统，对后端也有一定了解。"
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
                    bio: "对人工智能和机器学习有浓厚兴趣，参与过相关项目开发。"
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
                    bio: "后端开发工程师，有扎实的Java基础，熟悉主流框架和数据库。"
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
                    bio: "UI设计师，有良好的美学素养和用户体验意识，熟练使用设计工具。"
                }
            ],
            
            // 筛选选项
            educationOptions: ["本科","硕士","博士"],
            majorOptions: ["计算机科学与技术", "软件工程", "网络工程", "人工智能"],
            skillOptions: ["JavaScript", "Vue", "React", "TypeScript", "Node.js", "Python", "C++", "Java", "Spring Boot", "MySQL", "HTML/CSS", "UI设计", "Photoshop", "Figma", "机器学习"],
            experienceOptions: ["应届毕业生", "1年经验", "2年经验", "3年以上"],
            
            // 选中的筛选条件
            selectedFilters: {
                education: [],
                majors: [],
                skills: [],
                experience: []
            },
            
            // 新学生表单数据
            newStudent: {
                name: '',
                age: '',
                education: '',
                major: '',
                school: '',
                experience: '',
                skills: [],
                bio: ''
            },
            
            // 显示添加表单
            showAddForm: false
        }
    },
    
    computed: {
        // 根据筛选条件过滤学生
        filteredStudents() {
            return this.students.filter(student => {
                // 学历筛选
                if (this.selectedFilters.education.length > 0 && 
                    !this.selectedFilters.education.includes(student.education)) {
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
                
                return true;
            });
        }
    },
    
    methods: {
        // 重置筛选条件
        resetFilters() {
            this.selectedFilters.education = [];
            this.selectedFilters.majors = [];
            this.selectedFilters.skills = [];
            this.selectedFilters.experience = [];
        },
        
        // 添加新学生
        addStudent() {
            // 验证表单数据
            if (!this.newStudent.name || 
                !this.newStudent.age || 
                !this.newStudent.education ||
                !this.newStudent.major ||
                !this.newStudent.school ||
                !this.newStudent.experience ||
                !this.newStudent.bio) {
                alert('请填写完整信息');
                return;
            }
            
            // 创建新学生对象
            const student = {
                id: Date.now(), // 使用时间戳作为唯一ID
                name: this.newStudent.name,
                age: parseInt(this.newStudent.age),
                education: this.newStudent.education,
                major: this.newStudent.major,
                school: this.newStudent.school,
                experience: this.newStudent.experience,
                skills: [...this.newStudent.skills],
                bio: this.newStudent.bio
            };
            
            // 添加到学生列表
            this.students.unshift(student);
            
            // 更新专业选项（如果是新专业）
            if (!this.majorOptions.includes(this.newStudent.major)) {
                this.majorOptions.push(this.newStudent.major);
            }
            
            // 更新技能选项（如果有新技能）
            this.newStudent.skills.forEach(skill => {
                if (!this.skillOptions.includes(skill)) {
                    this.skillOptions.push(skill);
                }
            });
            
            // 重置表单
            this.resetNewStudentForm();
            
            // 关闭模态框
            this.showAddForm = false;
            
            alert('学生信息添加成功！');
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
                bio: ''
            };
        }
    }
});

// 挂载Vue应用
app.mount('#app');