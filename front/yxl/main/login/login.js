// 登录功能
document.getElementById('loginForm').addEventListener('submit', function(e) {
    e.preventDefault();
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const userType = document.getElementById('userType').value;
    if (!username || !password || !userType) {
        alert('请填写完整信息');
        return;
    }
    
    fetch('http://81.70.199.174:8081/login', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({username, password, userType})
    })
    .then(res => res.json())
    .then(data => {
        if (data.code === 200) {
            alert('登录成功');
            window.userInfo.setUserInfo(data.data);
            console.log(window.userInfo);
            
        } else {
            alert(data.message || '登录失败');
        }
    })
    .catch(error => {
        alert('网络错误');
    });
});