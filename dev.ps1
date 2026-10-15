param([int]$Port = 8095)
$ErrorActionPreference = 'Stop'
$goCommand = Get-Command go -ErrorAction SilentlyContinue
$goExe = if ($goCommand) { $goCommand.Source } else { Join-Path $env:ProgramFiles 'Go\bin\go.exe' }
if (-not (Test-Path -LiteralPath $goExe)) { throw 'Install Go for Windows first: https://go.dev/dl/' }
if (Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue) { throw "Port $Port is already in use. Stop the existing development server or use -Port 8096." }
$previous = @{}
foreach ($key in @('ADDR','FORUM_URL','COMMUNITY_DB','COMMUNITY_INSECURE_COOKIES','COMMUNITY_ADMIN_PASSWORD_FILE')) { $previous[$key] = [Environment]::GetEnvironmentVariable($key,'Process') }
Push-Location $PSScriptRoot
try {
 $env:ADDR = "127.0.0.1:$Port"
 $env:FORUM_URL = "http://localhost:$Port/community"
 $env:COMMUNITY_DB = Join-Path $PSScriptRoot 'data\development.db'
 $env:COMMUNITY_INSECURE_COOKIES = '1'
 Remove-Item Env:\COMMUNITY_ADMIN_PASSWORD_FILE -ErrorAction SilentlyContinue
 Write-Host "Windows development: http://localhost:$Port/community"
 Write-Host 'Stop with Ctrl+C. Restart after editing Go, HTML or CSS.'
 & $goExe run .
 if ($LASTEXITCODE -ne 0) { throw "Go exited with code $LASTEXITCODE" }
} finally {
 Pop-Location
 foreach ($key in $previous.Keys) { [Environment]::SetEnvironmentVariable($key,$previous[$key],'Process') }
}
