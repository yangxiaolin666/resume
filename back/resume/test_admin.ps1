# 管理端接口测试脚本
# 使用PowerShell执行

# 配置
$baseUrl = "http://localhost:8081"
$adminUser = "admin"
$adminPass = "admin123"

Write-Host "========== 管理端接口测试 ==========" -ForegroundColor Green

# 1. 登录获取Token
Write-Host "`n1. 登录管理员账号..." -ForegroundColor Yellow
$loginBody = @{
    username = $adminUser
    password = $adminPass
    user_type = "admin"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/login" -Method Post -Body $loginBody -ContentType "application/json"
    $token = $loginResponse.data.token
    Write-Host "✓ 登录成功！Token: $($token.Substring(0,20))..." -ForegroundColor Green
    
    # 设置请求头
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }
} catch {
    Write-Host "✗ 登录失败: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# 2. 查询学生列表
Write-Host "`n2. 查询学生列表..." -ForegroundColor Yellow
try {
    $studentsResponse = Invoke-RestMethod -Uri "$baseUrl/admin/searchStu?page=1&pageSize=5" -Method Get -Headers $headers
    Write-Host "✓ 查询成功！总数: $($studentsResponse.data.pagination.total)" -ForegroundColor Green
    Write-Host "  学生列表:" -ForegroundColor Cyan
    $studentsResponse.data.list | ForEach-Object {
        Write-Host "    - ID: $($_.id), 学号: $($_.stuId), 姓名: $($_.name)" -ForegroundColor White
    }
} catch {
    Write-Host "✗ 查询失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 3. 查询企业列表
Write-Host "`n3. 查询企业列表..." -ForegroundColor Yellow
try {
    $companiesResponse = Invoke-RestMethod -Uri "$baseUrl/admin/searchCom?page=1&pageSize=5" -Method Get -Headers $headers
    Write-Host "✓ 查询成功！总数: $($companiesResponse.data.pagination.total)" -ForegroundColor Green
    if ($companiesResponse.data.list.Count -gt 0) {
        Write-Host "  企业列表:" -ForegroundColor Cyan
        $companiesResponse.data.list | ForEach-Object {
            Write-Host "    - ID: $($_.id), 名称: $($_.comName), 地区: $($_.comRegion)" -ForegroundColor White
        }
    } else {
        Write-Host "  暂无企业数据" -ForegroundColor Gray
    }
} catch {
    Write-Host "✗ 查询失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 4. 测试更新学生（如果有学生数据）
if ($studentsResponse.data.list.Count -gt 0) {
    Write-Host "`n4. 测试更新学生数据..." -ForegroundColor Yellow
    $firstStudent = $studentsResponse.data.list[0]
    $updateBody = @{
        students = @(
            @{
                id = $firstStudent.id
                name = "$($firstStudent.name)_已更新"
            }
        )
    } | ConvertTo-Json -Depth 3
    
    try {
        $updateResponse = Invoke-RestMethod -Uri "$baseUrl/admin/updateStu" -Method Put -Body $updateBody -Headers $headers
        Write-Host "✓ 更新成功！成功数: $($updateResponse.data.successCount)" -ForegroundColor Green
        
        # 恢复原名称
        $restoreBody = @{
            students = @(
                @{
                    id = $firstStudent.id
                    name = $firstStudent.name
                }
            )
        } | ConvertTo-Json -Depth 3
        Invoke-RestMethod -Uri "$baseUrl/admin/updateStu" -Method Put -Body $restoreBody -Headers $headers | Out-Null
        Write-Host "  已恢复原数据" -ForegroundColor Gray
    } catch {
        Write-Host "✗ 更新失败: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# 5. 测试删除（创建测试数据后删除）
Write-Host "`n5. 测试删除功能（需要先导入测试数据）..." -ForegroundColor Yellow
Write-Host "  跳过自动删除测试，请手动测试" -ForegroundColor Gray

# 6. 测试Excel导入
Write-Host "`n6. 测试Excel导入..." -ForegroundColor Yellow
Write-Host "  请手动准备Excel文件并测试导入功能" -ForegroundColor Gray
Write-Host "  参考文档: Excel导入模板说明.md" -ForegroundColor Gray

Write-Host "`n========== 测试完成 ==========" -ForegroundColor Green
Write-Host "`n提示:" -ForegroundColor Cyan
Write-Host "1. Excel导入测试需要手动准备.xlsx文件" -ForegroundColor White
Write-Host "2. 使用Postman测试更方便" -ForegroundColor White
Write-Host "3. 详细使用说明请查看: 管理端使用指南.md" -ForegroundColor White

# 测试非管理员访问（验证权限）
Write-Host "`n========== 权限验证测试 ==========" -ForegroundColor Green
Write-Host "`n测试非管理员账号访问管理端接口..." -ForegroundColor Yellow

# 尝试使用学生账号登录并访问管理端接口
$studentLogin = @{
    username = "20230001"
    password = "20230001"
    user_type = "student"
} | ConvertTo-Json

try {
    Write-Host "  使用学生账号登录..." -ForegroundColor Gray
    $studentResponse = Invoke-RestMethod -Uri "$baseUrl/login" -Method Post -Body $studentLogin -ContentType "application/json" -ErrorAction Stop
    $studentToken = $studentResponse.data.token
    $studentHeaders = @{
        "Authorization" = "Bearer $studentToken"
    }
    
    Write-Host "  尝试访问管理端接口..." -ForegroundColor Gray
    try {
        Invoke-RestMethod -Uri "$baseUrl/admin/searchStu" -Method Get -Headers $studentHeaders -ErrorAction Stop
        Write-Host "✗ 权限验证失败！学生账号不应该能访问管理端接口" -ForegroundColor Red
    } catch {
        if ($_.Exception.Response.StatusCode -eq 403) {
            Write-Host "✓ 权限验证成功！学生账号被正确拒绝访问" -ForegroundColor Green
        } else {
            Write-Host "? 收到意外响应: $($_.Exception.Message)" -ForegroundColor Yellow
        }
    }
} catch {
    Write-Host "  学生账号未设置或登录失败，跳过权限测试" -ForegroundColor Gray
}

Write-Host "`n========== 所有测试完成 ==========" -ForegroundColor Green
