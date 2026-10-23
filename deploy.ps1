# One-click deploy script for backend
# Usage: after editing Java code, run .\deploy.ps1 in PowerShell
# Does: package -> upload -> restart -> verify

$ErrorActionPreference = "Stop"

# ===== Config (change server info here) =====
$server     = "182.92.210.198"
$user       = "root"
$pass       = '{9$R!Q7J#B+2{D-&EM&S=g~4q>'
$projectDir = "D:\resume\back\resume"
$jdkHome    = "C:\Program Files\Microsoft\jdk-17.0.20.101-hotspot"
$localJar   = "$projectDir\target\resume-0.0.1-SNAPSHOT.jar"
$remoteDir  = "/opt/resume"
$remoteJar  = "$remoteDir/resume-0.0.1-SNAPSHOT.jar"
# ============================================

function Step($n, $msg) { Write-Host "`n[$n/4] $msg" -ForegroundColor Cyan }

# 1. Package
Step 1 "Packaging jar..."
$env:JAVA_HOME = $jdkHome
$env:Path = "$jdkHome\bin;$env:Path"
Push-Location $projectDir
& mvn clean package -DskipTests 2>&1 | Select-Object -Last 8
Pop-Location
if (-not (Test-Path $localJar)) { Write-Host "Pack failed, jar not found" -ForegroundColor Red; exit 1 }
$jarMB = [math]::Round((Get-Item $localJar).Length / 1MB, 2)
Write-Host "    Pack done: $jarMB MB" -ForegroundColor Green

# 2. Upload
Step 2 "Uploading jar to server..."
Import-Module Posh-SSH -ErrorAction Stop
$secure = ConvertTo-SecureString $pass -AsPlainText -Force
$cred = New-Object System.Management.Automation.PSCredential($user, $secure)
$sftp = New-SFTPSession -ComputerName $server -Credential $cred -AcceptKey -ConnectionTimeout 30
Remove-SFTPItem -SessionId $sftp.SessionId -Path $remoteJar -Force -ErrorAction SilentlyContinue
Set-SFTPItem -SessionId $sftp.SessionId -Path $localJar -Destination $remoteDir
Remove-SFTPSession -SessionId $sftp.SessionId | Out-Null
Write-Host "    Upload done" -ForegroundColor Green

# 3. Restart
Step 3 "Restarting resume.service..."
$sess = New-SSHSession -ComputerName $server -Credential $cred -AcceptKey -ConnectionTimeout 30
$cmd = 'systemctl restart resume.service; sleep 25; echo STATUS: $(systemctl is-active resume.service); ss -tlnp | grep 8081 || echo NO_PORT'
$r = Invoke-SSHCommand -SessionId $sess.SessionId -Command $cmd -TimeOut 60
Write-Host $r.Output
Remove-SSHSession -SessionId $sess.SessionId | Out-Null

# 4. Verify
Step 4 "Verifying service..."
$headers = @{ "Content-Type" = "application/json" }
$body = '{"username":"_deploy_test_","password":"x","userType":"student"}'
$url = "http://${server}:8081/login"
try {
    $test = Invoke-RestMethod -Uri $url -Method POST -Headers $headers -Body $body -TimeoutSec 15
    $json = $test | ConvertTo-Json -Compress
    Write-Host "    Response: $json" -ForegroundColor Green
} catch {
    $code = 0
    if ($_.Exception.Response) { $code = [int]$_.Exception.Response.StatusCode }
    if ($code -eq 400) {
        Write-Host "    Service up, API returned 400 (check params)" -ForegroundColor Yellow
    } else {
        Write-Host "    Verify failed HTTP $code, maybe service not ready" -ForegroundColor Yellow
    }
}
Write-Host "`n=== Deploy done! ===" -ForegroundColor Green
Write-Host "Backend: http://${server}:8081" -ForegroundColor Cyan
Write-Host "API doc: http://${server}:8081/doc.html" -ForegroundColor Cyan
Write-Host "Log: SSH into server, run journalctl -u resume.service -f" -ForegroundColor Gray
