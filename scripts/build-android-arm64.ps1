$ErrorActionPreference = "Stop"

$ProjectRoot = Split-Path -Parent $PSScriptRoot
$OutputDir = Join-Path $ProjectRoot "dist"
$OutputBin = Join-Path $OutputDir "f50-web-arm64"

if (-not $env:ANDROID_NDK_HOME -and -not $env:ANDROID_NDK_ROOT) {
    throw "未设置 ANDROID_NDK_HOME 或 ANDROID_NDK_ROOT"
}

$NdkRoot = $env:ANDROID_NDK_HOME
if (-not $NdkRoot) {
    $NdkRoot = $env:ANDROID_NDK_ROOT
}

$Toolchain = Join-Path $NdkRoot "toolchains\llvm\prebuilt\windows-x86_64\bin"
$Compiler = Join-Path $Toolchain "aarch64-linux-android24-clang.cmd"
if (-not (Test-Path $Compiler)) {
    $Compiler = Join-Path $Toolchain "aarch64-linux-android24-clang"
}
if (-not (Test-Path $Compiler)) {
    throw "未找到 Android arm64 编译器: $Compiler"
}

New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

$env:CGO_ENABLED = "1"
$env:GOOS = "android"
$env:GOARCH = "arm64"
$env:CC = $Compiler

Push-Location $ProjectRoot
try {
    go build -trimpath -o $OutputBin .
}
finally {
    Pop-Location
}
