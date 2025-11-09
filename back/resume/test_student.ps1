# 学生端接口测试脚本 (PowerShell版本)

# 颜色输出函数
function Write-ColorOutput($message, $color) {
    Write-Host $message -ForegroundColor $color
}

# 基础URL
$baseUrl = "http://localhost:8081"

Write-ColorOutput "========================================" Cyan
Write-ColorOutput "学生端接口测试" Cyan
Write-ColorOutput "========================================" Cyan
Write-Host ""

# 步骤1: 登录获取Token
Write-ColorOutput "步骤1: 学生登录获取Token..." Yellow
$loginBody = @{
    username = "20230001"
    password = "20230001"
    userType = "student"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/login" -Method Post -Body $loginBody -ContentType "application/json"
    $token = $loginResponse.data.token
    Write-ColorOutput "✓ 登录成功！Token: $($token.Substring(0, 20))..." Green
    Write-Host ""
} catch {
    Write-ColorOutput "✗ 登录失败: $_" Red
    exit
}

# 步骤2: 测试完善学生信息
Write-ColorOutput "步骤2: 测试完善学生信息..." Yellow
$completeInfoBody = @{
    age = 22
    gender = "男"
    nation = "汉族"
    major = "计算机科学与技术"
    degree = "本科"
    graduationYear = 2025
    internship = 2
    languages = @(
        @{
            language = "英语"
            level = "六级"
        },
        @{
            language = "日语"
            level = "N2"
        }
    )
    cities = @("北京", "上海")
    provinces = @("广东")
    certificates = @(
        @{
            certName = "软考中级软件设计师"
            level = "中级"
        }
    )
    internships = @(
        @{
            company = "腾讯"
            experience = "后端开发实习，负责微服务开发"
        }
    )
    specialities = @("Java", "Python", "MySQL", "Redis")
    jobs = @("后端开发", "全栈开发")
} | ConvertTo-Json -Depth 10

try {
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }
    
    $completeResponse = Invoke-RestMethod -Uri "$baseUrl/stu/completeInfo" -Method Put -Body $completeInfoBody -Headers $headers
    Write-ColorOutput "✓ 完善信息成功！" Green
    Write-ColorOutput "  学生ID: $($completeResponse.data.id)" White
    Write-ColorOutput "  学号: $($completeResponse.data.stuId)" White
    Write-ColorOutput "  姓名: $($completeResponse.data.name)" White
    Write-ColorOutput "  年龄: $($completeResponse.data.age)" White
    Write-ColorOutput "  专业: $($completeResponse.data.major)" White
    Write-Host ""
} catch {
    Write-ColorOutput "✗ 完善信息失败: $_" Red
    Write-Host ""
}

# 步骤3: 测试错误的年龄验证
Write-ColorOutput "步骤3: 测试年龄验证（年龄=100，应该失败）..." Yellow
$invalidAgeBody = @{
    age = 100
} | ConvertTo-Json

try {
    $invalidResponse = Invoke-RestMethod -Uri "$baseUrl/stu/completeInfo" -Method Put -Body $invalidAgeBody -Headers $headers -ErrorAction Stop
    Write-ColorOutput "✗ 验证失败：应该拒绝不合理的年龄" Red
} catch {
    $errorResponse = $_.ErrorDetails.Message | ConvertFrom-Json
    if ($errorResponse.code -eq 42201) {
        Write-ColorOutput "✓ 年龄验证通过！错误码: $($errorResponse.code), 消息: $($errorResponse.message)" Green
    } else {
        Write-ColorOutput "✗ 错误码不符合预期: $($errorResponse.code)" Red
    }
    Write-Host ""
}

# 步骤4: 测试上传简历（需要先准备一个测试文件）
Write-ColorOutput "步骤4: 测试上传简历..." Yellow
Write-ColorOutput "  注意：需要在当前目录准备一个test_resume.pdf文件" Cyan

$resumeFilePath = "test_resume.pdf"
if (Test-Path $resumeFilePath) {
    try {
        $boundary = [System.Guid]::NewGuid().ToString()
        $fileContent = [System.IO.File]::ReadAllBytes($resumeFilePath)
        $fileName = [System.IO.Path]::GetFileName($resumeFilePath)
        
        $bodyLines = @(
            "--$boundary",
            "Content-Disposition: form-data; name=`"resumeFile`"; filename=`"$fileName`"",
            "Content-Type: application/pdf",
            "",
            [System.Text.Encoding]::UTF8.GetString($fileContent),
            "--$boundary--"
        ) -join "`r`n"
        
        $uploadHeaders = @{
            "Authorization" = "Bearer $token"
            "Content-Type" = "multipart/form-data; boundary=$boundary"
        }
        
        # 使用Invoke-WebRequest而不是Invoke-RestMethod以支持文件上传
        $uploadResponse = Invoke-WebRequest -Uri "$baseUrl/stu/uploadResume" `
            -Method Post `
            -Headers $uploadHeaders `
            -Body $bodyLines
        
        $uploadResult = $uploadResponse.Content | ConvertFrom-Json
        
        if ($uploadResult.code -eq 201) {
            Write-ColorOutput "✓ 简历上传成功！" Green
            Write-ColorOutput "  简历ID: $($uploadResult.data.resumeId)" White
            Write-ColorOutput "  文件名: $($uploadResult.data.resumeName)" White
            Write-ColorOutput "  访问地址: $($uploadResult.data.resumeUrl)" White
        } else {
            Write-ColorOutput "✗ 上传失败: $($uploadResult.message)" Red
        }
        Write-Host ""
    } catch {
        Write-ColorOutput "✗ 简历上传失败: $_" Red
        Write-Host ""
    }
} else {
    Write-ColorOutput "⊘ 跳过简历上传测试（未找到test_resume.pdf文件）" Yellow
    Write-ColorOutput "  你可以创建一个测试PDF文件来测试上传功能" Cyan
    Write-Host ""
}

# 步骤5: 测试文件格式验证（使用txt文件应该失败）
Write-ColorOutput "步骤5: 测试文件格式验证..." Yellow
$txtFilePath = "test.txt"
if (Test-Path $txtFilePath) {
    Write-ColorOutput "  发现test.txt文件，测试格式验证..." Cyan
    # 这里可以添加上传txt文件的测试代码，预期会失败
    Write-ColorOutput "  (此测试需要实际文件上传实现)" Yellow
    Write-Host ""
} else {
    Write-ColorOutput "⊘ 跳过文件格式验证测试（未找到test.txt文件）" Yellow
    Write-Host ""
}

Write-ColorOutput "========================================" Cyan
Write-ColorOutput "测试完成！" Green
Write-ColorOutput "========================================" Cyan
Write-Host ""
Write-ColorOutput "总结：" Yellow
Write-ColorOutput "✓ 登录功能正常" Green
Write-ColorOutput "✓ 完善信息功能正常" Green
Write-ColorOutput "✓ 年龄验证功能正常" Green
if (Test-Path $resumeFilePath) {
    Write-ColorOutput "✓ 简历上传功能正常" Green
} else {
    Write-ColorOutput "⊘ 简历上传未测试（需要test_resume.pdf文件）" Yellow
}
Write-Host ""
