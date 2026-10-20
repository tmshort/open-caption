@echo off
rem Copyright the open-caption authors
rem SPDX-License-Identifier: Apache-2.0
rem Builds open-caption.exe (see build.ps1 for the options). Works from cmd.exe, PowerShell or a double-click,
rem whatever PowerShell's script policy is.
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0build.ps1" %*
set RC=%ERRORLEVEL%
exit /b %RC%
