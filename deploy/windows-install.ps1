# Windows: run open-caption automatically at startup, using Task Scheduler.
# Run this in an ADMINISTRATOR PowerShell window, from the folder containing open-caption.exe:
#
#     powershell -ExecutionPolicy Bypass -File windows-install.ps1
#
# Remove it again with:  schtasks /Delete /TN open-caption /F
$dir  = Split-Path -Parent $MyInvocation.MyCommand.Path
$exe  = Join-Path $dir 'open-caption.exe'
$data = Join-Path $dir 'data'
if (-not (Test-Path $exe)) { Write-Error "open-caption.exe was not found in $dir"; exit 1 }
New-Item -ItemType Directory -Force -Path $data | Out-Null
$exeArgs = "--data `"$data`" --no-browser --log-file `"$(Join-Path $dir 'open-caption.log')`""
$action  = New-ScheduledTaskAction -Execute $exe -Argument $exeArgs -WorkingDirectory $dir
$trigger = New-ScheduledTaskTrigger -AtStartup
$settings = New-ScheduledTaskSettingsSet -RestartCount 5 -RestartInterval (New-TimeSpan -Minutes 1) -StartWhenAvailable
Register-ScheduledTask -TaskName 'open-caption' -Action $action -Trigger $trigger -Settings $settings -RunLevel Highest -Force | Out-Null
Start-ScheduledTask -TaskName 'open-caption'
Write-Host "Installed. Passwords are in $data\config.yaml"
Write-Host "Also allow open-caption.exe through Windows Firewall on PRIVATE networks."
