@echo off
setlocal enabledelayedexpansion
cd /d "%~dp0"

set "BFF=teamgramd\etc2\bff.yaml"
set "IP_FILE=.public_ip"
set "SECRET_FILE=.turn_secret"
set "ENV_FILE=.env"
set "PROFILE_FILE=.env_profile"

rem --- Infrastructure profile (interactive) -----------------------------
rem All three share the same core (kafka/etcd/redis/mysql/minio); they differ
rem in memory limits and the logging/tracing/metrics services.
set "DEFAULT_PROFILE=default"
if exist "%PROFILE_FILE%" set /p DEFAULT_PROFILE=<"%PROFILE_FILE%"
echo Choose an infrastructure profile:
echo   [1] minimal  - core only + strict memory limits     approx 2-4 GB RAM
echo   [2] default  - core only, no memory limits          approx 4-8 GB RAM
echo   [3] full     - core + logging / tracing / metrics   approx 16 GB RAM
set /p "PROFILE_CHOICE=Profile [%DEFAULT_PROFILE%]: "
if not defined PROFILE_CHOICE set "PROFILE_CHOICE=%DEFAULT_PROFILE%"
set "ENV_PROFILE="
if /i "%PROFILE_CHOICE%"=="1"       set "ENV_PROFILE=min"
if /i "%PROFILE_CHOICE%"=="min"     set "ENV_PROFILE=min"
if /i "%PROFILE_CHOICE%"=="minimal" set "ENV_PROFILE=min"
if /i "%PROFILE_CHOICE%"=="2"       set "ENV_PROFILE=default"
if /i "%PROFILE_CHOICE%"=="default" set "ENV_PROFILE=default"
if /i "%PROFILE_CHOICE%"=="3"       set "ENV_PROFILE=full"
if /i "%PROFILE_CHOICE%"=="full"    set "ENV_PROFILE=full"
if not defined ENV_PROFILE (
  echo [ERROR] invalid profile: %PROFILE_CHOICE%
  pause
  exit /b 1
)
set "ENV_COMPOSE=docker-compose-env-%ENV_PROFILE%.yaml"
> "%PROFILE_FILE%" echo %ENV_PROFILE%
echo [cfg] infrastructure profile = %ENV_PROFILE% : %ENV_COMPOSE%

rem --- Public address (interactive) -------------------------------------
rem Public IP/host that remote clients use to reach this server. Baked into
rem the MTProto + VoIP/TURN config so chats AND calls work globally.
set "DEFAULT_IP="
if exist "%IP_FILE%" set /p DEFAULT_IP=<"%IP_FILE%"
if defined DEFAULT_IP (
  set /p "PUBLIC_IP=Public server IP/host [%DEFAULT_IP%]: "
) else (
  set /p "PUBLIC_IP=Public server IP/host: "
)
if not defined PUBLIC_IP set "PUBLIC_IP=%DEFAULT_IP%"
if not defined PUBLIC_IP (
  echo [ERROR] public IP/host is required.
  pause
  exit /b 1
)
> "%IP_FILE%" echo %PUBLIC_IP%

rem --- TURN secret (generated once, reused) -----------------------------
set "TURN_SECRET="
if exist "%SECRET_FILE%" set /p TURN_SECRET=<"%SECRET_FILE%"
if not defined TURN_SECRET (
  for /f "usebackq delims=" %%i in (`powershell -NoProfile -Command "[guid]::NewGuid().ToString('N')"`) do set "TURN_SECRET=%%i"
  > "%SECRET_FILE%" echo !TURN_SECRET!
)

rem --- compose env (consumed by the coturn service) ---------------------
> "%ENV_FILE%" echo PUBLIC_IP=%PUBLIC_IP%
>> "%ENV_FILE%" echo TURN_SECRET=!TURN_SECRET!

rem --- keep a pristine copy; restored after the build so the rendered TURN
rem     secret is never left on disk / committed by accident ----------------
set "BFF_PRISTINE=%TEMP%\owpengram_bff_pristine.yaml"
copy /Y "%BFF%" "%BFF_PRISTINE%" >nul

rem --- bake public address + TURN secret into the server config ---------
powershell -NoProfile -ExecutionPolicy Bypass -Command "$ip='%PUBLIC_IP%'; $sec='!TURN_SECRET!'; $f=(Resolve-Path '%BFF%').Path; $enc=New-Object System.Text.UTF8Encoding($false); $t=[System.IO.File]::ReadAllText($f); $t=$t.TrimStart([char]0xFEFF); $t=$t -replace '(?m)^(\s*Ip:\s*).*$', ('${1}'+$ip); $t=$t -replace '(?m)^(\s*Password:\s*).*$', ('${1}\"'+$sec+'\"'); [System.IO.File]::WriteAllText($f,$t,$enc)"
if %ERRORLEVEL% neq 0 (
  echo [ERROR] failed to update %BFF%
  call :restore_bff
  pause
  exit /b 1
)

rem --- render coturn config from template (public IP + TURN secret) -----
if not exist coturn mkdir coturn
powershell -NoProfile -ExecutionPolicy Bypass -Command "$ip='%PUBLIC_IP%'; $sec='!TURN_SECRET!'; $base=(Get-Location).Path; $tpl=Join-Path $base 'coturn\turnserver.conf.template'; $out=Join-Path $base 'coturn\turnserver.conf'; $enc=New-Object System.Text.UTF8Encoding($false); $t=[System.IO.File]::ReadAllText($tpl); $t=$t -replace '__PUBLIC_IP__',$ip -replace '__TURN_SECRET__',$sec; [System.IO.File]::WriteAllText($out,$t,$enc)"
if %ERRORLEVEL% neq 0 (
  echo [ERROR] failed to render coturn\turnserver.conf
  pause
  exit /b %ERRORLEVEL%
)
echo [cfg] public address = %PUBLIC_IP%; TURN relay configured.

rem --- Windows firewall (best-effort; requires admin) ------------------
for %%R in ("owpengram 10443" "owpengram turn 3478 udp" "owpengram turn 3478 tcp" "owpengram turn media") do netsh advfirewall firewall delete rule name=%%R >nul 2>&1
netsh advfirewall firewall add rule name="owpengram 10443" dir=in action=allow protocol=TCP localport=10443 >nul 2>&1
netsh advfirewall firewall add rule name="owpengram turn 3478 udp" dir=in action=allow protocol=UDP localport=3478 >nul 2>&1
netsh advfirewall firewall add rule name="owpengram turn 3478 tcp" dir=in action=allow protocol=TCP localport=3478 >nul 2>&1
netsh advfirewall firewall add rule name="owpengram turn media" dir=in action=allow protocol=UDP localport=49160-49200 >nul 2>&1

echo.
rem When not 'full', remove leftover observability-only containers from a previous
rem full run (safe: only these fixed names, never the core or app container).
if /i not "%ENV_PROFILE%"=="full" (
  for %%C in (jaeger grafana prometheus kibana elasticsearch filebeat go-stash node-exporter) do docker rm -f %%C >nul 2>&1
)
echo [1/3] infrastructure "%ENV_PROFILE%" : docker compose -f %ENV_COMPOSE% up -d
docker compose -f "%ENV_COMPOSE%" up -d
if %ERRORLEVEL% neq 0 (
  echo [ERROR] env stack failed
  call :restore_bff
  pause
  exit /b 1
)

echo.
echo [1.5/3] waiting for MySQL, then applying DB migrations (idempotent)
set /a _myw=0
:wait_mysql
docker exec mysql mysql -uteamgram -pteamgram teamgram -N -e "SELECT 1" >nul 2>&1
if not errorlevel 1 goto do_migrations
set /a _myw+=1
if !_myw! gtr 60 ( echo [WARN] MySQL not ready ~120s - skipping migrations & goto after_migrations )
timeout /t 2 >nul
goto wait_mysql
:do_migrations
docker exec mysql mysql -uteamgram -pteamgram teamgram -e "CREATE TABLE IF NOT EXISTS schema_migrations (name VARCHAR(128) NOT NULL PRIMARY KEY, applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);" >nul 2>&1
rem Apply each migrate-*.sql in date order, only if not already recorded. --force tolerates
rem "already applied" duplicate-column/key errors on a pre-existing DB, so nothing breaks.
for /f "delims=" %%F in ('dir /b /on "teamgramd\deploy\sql\migrate-*.sql" 2^>nul') do (
  set "_done="
  for /f "usebackq delims=" %%C in (`docker exec mysql mysql -uteamgram -pteamgram teamgram -N -e "SELECT COUNT(1) FROM schema_migrations WHERE name='%%F'" 2^>nul`) do set "_done=%%C"
  if not "!_done!"=="1" (
    echo   [migrate] %%F
    docker exec -i mysql mysql --force -uteamgram -pteamgram teamgram < "teamgramd\deploy\sql\%%F" >nul 2>&1
    docker exec mysql mysql -uteamgram -pteamgram teamgram -e "INSERT IGNORE INTO schema_migrations (name) VALUES ('%%F');" >nul 2>&1
  )
)
echo [OK] DB migrations applied.
:after_migrations

echo.
echo [2/3] docker compose up -d --build  (core server)
rem --build so the edited config (public address) is baked into the image.
docker compose up -d --build
set "BUILD_ERR=%ERRORLEVEL%"
rem revert bff.yaml to placeholders now that the image is built with the secret
call :restore_bff
if %BUILD_ERR% neq 0 (
  echo [ERROR] app stack failed
  pause
  exit /b %BUILD_ERR%
)

echo.
echo [3/3] coturn (calls relay) - best-effort, never blocks the core server
docker compose -f docker-compose-turn.yaml up -d
if %ERRORLEVEL% neq 0 (
  echo [WARN] coturn failed to start - calls relay unavailable; core server is fine.
)

echo.
echo [OK] Server started (public address: %PUBLIC_IP%).
echo      Also open these ports in the VPS PROVIDER firewall:
echo        TCP 10443        - MTProto (login / chats / media)
echo        UDP+TCP 3478     - TURN/STUN control (calls)
echo        UDP 49160-49200  - TURN media relay (calls)
exit /b 0

:restore_bff
rem restore the tracked bff.yaml from the pristine copy (placeholders), so the
rem rendered public IP + TURN secret are never left in the working tree.
if defined BFF_PRISTINE if exist "%BFF_PRISTINE%" (
  copy /Y "%BFF_PRISTINE%" "%BFF%" >nul
  del /Q "%BFF_PRISTINE%" >nul
  echo [cfg] bff.yaml reverted to placeholders ^(secret not left on disk^).
)
exit /b 0
