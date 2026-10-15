param([int]$Port = 8095)
$ErrorActionPreference = 'Stop'
$listener = Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue
if (-not $listener) { Write-Host 'Development server is already stopped.'; exit }
foreach ($serverId in ($listener.OwningProcess | Select-Object -Unique)) {
 $process = Get-Process -Id $serverId
 if ($process.ProcessName -ne 'salafinetwork' -or $process.Path -notlike "$env:TEMP\go-build*\*\exe\salafinetwork.exe") { throw "Port $Port belongs to a different process; leaving it alone." }
 Stop-Process -Id $serverId
}
Write-Host 'Windows development server stopped.'
