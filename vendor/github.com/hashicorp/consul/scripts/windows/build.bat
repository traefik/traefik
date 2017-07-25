@echo off

setlocal

if not exist %1 exit /B 1
cd %1

:: Get the git commit
set _GIT_COMMIT_FILE=%TEMP%\consul-git_commit.txt
set _GIT_DIRTY_FILE=%TEMP%\consul-git_dirty.txt
set _GIT_DESCRIBE_FILE=%TEMP%\consul-git_describe.txt

set _NUL_CMP_FILE=%TEMP%\consul-nul_cmp.txt
type NUL >%_NUL_CMP_FILE%

git rev-parse HEAD >%_GIT_COMMIT_FILE%
set /p _GIT_COMMIT=<%_GIT_COMMIT_FILE%
del /F "%_GIT_COMMIT_FILE%" 2>NUL

set _GIT_DIRTY=
git status --porcelain >%_GIT_DIRTY_FILE%
fc %_GIT_DIRTY_FILE% %_NUL_CMP_FILE% >NUL
if errorlevel 1 set _GIT_DIRTY=+CHANGES
del /F "%_GIT_DIRTY_FILE%" 2>NUL
del /F "%_NUL_CMP_FILE%" 2>NUL

git describe --tags >%_GIT_DESCRIBE_FILE%
set /p _GIT_DESCRIBE=<%_GIT_DESCRIBE_FILE%
del /F "%_GIT_DESCRIBE_FILE%" 2>NUL

:: Install dependencies
echo --^> Installing dependencies to speed up builds...
go get .\...

:: Build!
echo --^> Building...
go build^
 -ldflags "-X main.GitCommit=%_GIT_COMMIT%%_GIT_DIRTY% -X main.GitDescribe=%_GIT_DESCRIBE%"^
 -v^
 -o bin\consul.exe .
if errorlevel 1 exit /B 1
copy /B /Y bin\consul.exe %GOPATH%\bin\consul.exe >NUL
