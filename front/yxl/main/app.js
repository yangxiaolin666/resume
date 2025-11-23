// 全局用户信息存储
window.userInfo = {
    data: null,
    token: null,
    refreshToken: null,
    
    // 设置用户信息
    setUserInfo: function(userData) {
        this.data = userData;
        this.token = userData.token;
        this.refreshToken = userData.refreshToken;
        console.log('用户信息已保存到app.js');
    },
    
    
    clearUserInfo: function() {
        this.data = null;
        this.token = null;
        this.refreshToken = null;
        console.log('用户信息已清除');
    }
};