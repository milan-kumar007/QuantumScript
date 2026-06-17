# QuantumScript Global Installer for Windows

$InstallDir = "$env:USERPROFILE\QuantumScript\bin"
$QsExe = "$InstallDir\qs.exe"

Write-Host "Installing QuantumScript CLI..." -ForegroundColor Cyan

# 1. Create directory
if (!(Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

# 2. Compile to directory
Write-Host "Compiling engine..."
go build -o $QsExe cli/main.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "Compilation failed!" -ForegroundColor Red
    exit 1
}

# 3. Add to User PATH if not already there
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notmatch [regex]::Escape($InstallDir)) {
    Write-Host "Adding $InstallDir to system PATH..."
    $NewPath = $UserPath + ";$InstallDir"
    [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
    Write-Host "Added to PATH. Please restart your terminal to use 'qs' globally." -ForegroundColor Yellow
} else {
    Write-Host "PATH is already configured."
}

Write-Host "QuantumScript installed successfully!" -ForegroundColor Green
Write-Host "Try running: qs version" -ForegroundColor Cyan
