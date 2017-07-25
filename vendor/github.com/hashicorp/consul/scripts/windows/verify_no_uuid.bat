@echo off

setlocal

if not exist %1\consul\state\state_store.go exit /B 1
if not exist %1\consul\fsm.go exit /B 1

findstr /R GenerateUUID %1\consul\state\state_store.go 1>nul
if not %ERRORLEVEL% EQU 1 exit /B 1

findstr GenerateUUID %1\consul\fsm.go 1>nul
if not %ERRORLEVEL% EQU 1 exit /B 1

exit /B 0
