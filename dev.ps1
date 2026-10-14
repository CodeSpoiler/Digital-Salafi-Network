$ErrorActionPreference = 'Stop'
Push-Location $PSScriptRoot
try { wsl.exe -d Ubuntu -- bash ./scripts/dev.sh } finally { Pop-Location }
