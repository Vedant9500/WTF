$ErrorActionPreference = "Stop"

Write-Host "Checking for Go..."

if (Get-Command go -ErrorAction SilentlyContinue) {
    Write-Host "Go found in PATH"
}
elseif (Test-Path "C:\Program Files\Go\bin\go.exe") {
    Write-Host "Go found in Program Files"
    $env:Path = "C:\Program Files\Go\bin;$env:Path"
}
elseif (Test-Path "C:\Go\bin\go.exe") {
    Write-Host "Go found in C:\Go"
    $env:Path = "C:\Go\bin;$env:Path"
}
elseif (Test-Path "C:\Users\sarth\Downloads\go1.23.1.windows-386\go\bin\go.exe") {
    Write-Host "Go found in User Downloads"
    $env:Path = "C:\Users\sarth\Downloads\go1.23.1.windows-386\go\bin;$env:Path"
}
else {
    Write-Error "Go not found. Please install Go."
    exit 1
}

go version
go mod tidy
go build -o wtf.exe ./cmd/wtf

if (Test-Path "wtf.exe") {
    Write-Host "Build success!"
    .\wtf.exe --help
}
