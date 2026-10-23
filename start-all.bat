@echo off
setlocal EnableExtensions EnableDelayedExpansion
cd /d "%~dp0"

set "ROOT=%~dp0"
set "GO_DIR=%ROOT%back\resume-go"
set "FRONT_DIR=%ROOT%front\yxl\recruit-project"
set "ADMIN_DIR=%ROOT%front\ljk\adminPanel"
set "MYSQL_BIN=C:\Program Files\MySQL\MySQL Server 8.0\bin\mysqld.exe"
set "MYSQL_CLIENT=C:\Program Files\MySQL\MySQL Server 8.0\bin\mysql.exe"
set "MYSQL_HOME=C:\Program Files\MySQL\MySQL Server 8.0"
set "MYSQL_DATA=C:\Users\HP\AppData\Local\resume-mysql\data"
set "BACKEND_PORT=8081"
set "FRONT_PORT=5173"
set "ADMIN_PORT=5174"

echo ============================================
echo   Resume Recruit System - One Click Start
echo ============================================
echo.

rem ---------- 1. database ----------
set "DB_PORT="
call :port_open 3306
if not errorlevel 1 set "DB_PORT=3306"
if defined DB_PORT goto db_ready

call :port_open 3307
if not errorlevel 1 set "DB_PORT=3307"
if defined DB_PORT goto db_ready

if not exist "%MYSQL_DATA%" goto db_missing
echo [1/4] MySQL not detected, starting local test instance on 3307...
start "resume-mysql" /min "%MYSQL_BIN%" --no-defaults --basedir="%MYSQL_HOME%" --datadir="%MYSQL_DATA%" --port=3307 --bind-address=127.0.0.1 --console
call :wait_port 3307 25
if errorlevel 1 goto db_missing
set "DB_PORT=3307"
goto db_ready

:db_missing
echo [ERROR] No MySQL listening on 3306 or 3307.
echo         Run this in an Administrator terminal: net start MySQL80
echo         Then double-click this file again.
pause
exit /b 1

:db_ready
echo [1/4] MySQL ready on port %DB_PORT%

rem ensure database schema and demo accounts exist (idempotent)
if exist "%MYSQL_CLIENT%" "%MYSQL_CLIENT%" --host=127.0.0.1 --port=%DB_PORT% --user=root --password=Wangtianye123 --default-character-set=utf8mb4 --database=resume --execute="source %GO_DIR%\schema.sql" >nul 2>&1
if exist "%MYSQL_CLIENT%" "%MYSQL_CLIENT%" --host=127.0.0.1 --port=%DB_PORT% --user=root --password=Wangtianye123 --default-character-set=utf8mb4 --database=resume --execute="source %GO_DIR%\seed_accounts.sql" >nul 2>&1
if exist "%MYSQL_CLIENT%" "%MYSQL_CLIENT%" --host=127.0.0.1 --port=%DB_PORT% --user=root --password=Wangtianye123 --default-character-set=utf8mb4 --database=resume --execute="source %GO_DIR%\seed_statistics.sql" >nul 2>&1

rem ---------- 2. Go backend ----------
call :port_open %BACKEND_PORT%
if not errorlevel 1 goto backend_ready

echo [2/4] Starting Go backend...
set "DB_HOST=127.0.0.1"
set "DB_PORT=%DB_PORT%"
start "resume-backend" /D "%GO_DIR%" cmd /k "go run ."
call :wait_port %BACKEND_PORT% 90
if errorlevel 1 goto backend_failed

:backend_ready
echo [2/4] Backend ready: http://localhost:%BACKEND_PORT%/
goto front_start

:backend_failed
echo [ERROR] Backend failed to start. Check the resume-backend window.
pause
exit /b 1

rem ---------- 3. frontend ----------
:front_start
call :port_open %FRONT_PORT%
if not errorlevel 1 goto front_ready

echo [3/4] Starting frontend...
if exist "%FRONT_DIR%\node_modules" goto front_deps_ok
echo       Installing frontend dependencies for the first run...
pushd "%FRONT_DIR%"
call npm install
popd

:front_deps_ok
set "VITE_API_BASE_URL=http://127.0.0.1:%BACKEND_PORT%"
set "VITE_ADMIN_URL=http://127.0.0.1:%ADMIN_PORT%"
start "resume-frontend" /D "%FRONT_DIR%" cmd /k "npm run dev -- --host 127.0.0.1 --port %FRONT_PORT% --strictPort"
call :wait_port %FRONT_PORT% 90
if errorlevel 1 goto front_failed

:front_ready
echo [3/4] Frontend ready: http://127.0.0.1:%FRONT_PORT%/
goto admin_start

:front_failed
echo [ERROR] Frontend failed to start. Check the resume-frontend window.
pause
exit /b 1

rem ---------- 4. admin panel ----------
:admin_start
call :port_open %ADMIN_PORT%
if not errorlevel 1 goto admin_ready

echo [4/4] Starting admin panel...
if exist "%ADMIN_DIR%\node_modules" goto admin_deps_ok
echo       Installing admin panel dependencies for the first run...
pushd "%ADMIN_DIR%"
call npm install
popd

:admin_deps_ok
set "VITE_API_BASE_URL=http://127.0.0.1:%BACKEND_PORT%"
set "VITE_LOGIN_URL=http://127.0.0.1:%FRONT_PORT%"
start "resume-admin" /D "%ADMIN_DIR%" cmd /k "npm run dev -- --host 127.0.0.1 --port %ADMIN_PORT% --strictPort"
call :wait_port %ADMIN_PORT% 90
if errorlevel 1 goto admin_failed

:admin_ready
echo [4/4] Admin panel ready: http://127.0.0.1:%ADMIN_PORT%/admin/
goto done

:admin_failed
echo [ERROR] Admin panel failed to start. Check the resume-admin window.
pause
exit /b 1

rem ---------- done ----------
:done
echo.
echo ============================================
echo   Project is running
echo   Frontend : http://127.0.0.1:%FRONT_PORT%/
echo   Admin UI : http://127.0.0.1:%ADMIN_PORT%/admin/
echo   Backend  : http://localhost:%BACKEND_PORT%/
echo.
echo   Admin    : g1 / g1
echo   Student  : x1 ~ x10 / 密码同账号名
echo   Company  : q1 ~ q10 / 密码同账号名
echo ============================================
echo.
if not defined NO_BROWSER start "" "http://127.0.0.1:%FRONT_PORT%/"
echo Browser opened. Press any key to close this window.
echo (Backend and frontend windows keep running.)
if not defined NO_PAUSE pause >nul
exit /b 0

rem ---------- helpers ----------
:port_open
netstat -ano | findstr /C:":%1 " | findstr /C:"LISTENING" >nul 2>&1
if errorlevel 1 exit /b 1
exit /b 0

:wait_port
set /a WAIT_LEFT=%2
:wait_loop
call :port_open %1
if not errorlevel 1 exit /b 0
set /a WAIT_LEFT-=1
if !WAIT_LEFT! LEQ 0 exit /b 1
ping -n 2 127.0.0.1 >nul
goto wait_loop
