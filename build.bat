@echo off
REM Build script for service-center (Windows)

setlocal enabledelayedexpansion

set VERSION=dev
set BUILD_DIR=_output\bin
set GOOS=
set GOARCH=

:parse_args
if "%~1"=="" goto build
if /i "%~1"=="-h" goto usage
if /i "%~1"=="--help" goto usage
if /i "%~1"=="-v" (
    set VERSION=%~2
    shift
    shift
    goto parse_args
)
if /i "%~1"=="--version" (
    set VERSION=%~2
    shift
    shift
    goto parse_args
)
if /i "%~1"=="server" (
    set BUILD_SERVER=1
    shift
    goto parse_args
)
if /i "%~1"=="client" (
    set BUILD_CLIENT=1
    shift
    goto parse_args
)
shift
goto parse_args

:usage
echo Build script for service-center
echo.
echo Usage: %~nx0 [OPTIONS] [BINARIES...]
echo.
echo OPTIONS:
echo     -h, --help              Show this help message
echo     -v, --version VERSION   Set version (default: dev)
echo.
echo BINARIES:
echo     server                  Build server binary
echo     client                  Build client binary
echo     (none)                  Build all binaries
echo.
echo EXAMPLES:
echo     %~nx0                      # Build all binaries
echo     %~nx0 server               # Build only server
echo     %~nx0 -v 1.0.0 server client
goto :eof

:build
REM If no targets specified, build all
if not defined BUILD_SERVER if not defined BUILD_CLIENT (
    set BUILD_SERVER=1
    set BUILD_CLIENT=1
)

REM Detect current platform
for /f "tokens=*" %%i in ('go env GOOS') do set GOOS=%%i
for /f "tokens=*" %%i in ('go env GOARCH') do set GOARCH=%%i

set OUTPUT_DIR=%BUILD_DIR%\%GOOS%\%GOARCH%
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

if defined BUILD_SERVER (
    echo Building server for %GOOS%/%GOARCH%...
    set CGO_ENABLED=0
    go build -ldflags "-w -s -X main.Version=%VERSION%" -o "%OUTPUT_DIR%\server.exe" .\cmd\server
    if !errorlevel! neq 0 (
        echo Error building server
        exit /b 1
    )
    echo   ✓ Built: %OUTPUT_DIR%\server.exe
)

if defined BUILD_CLIENT (
    echo Building client for %GOOS%/%GOARCH%...
    set CGO_ENABLED=0
    go build -ldflags "-w -s -X main.Version=%VERSION%" -o "%OUTPUT_DIR%\client.exe" .\cmd\client
    if !errorlevel! neq 0 (
        echo Error building client
        exit /b 1
    )
    echo   ✓ Built: %OUTPUT_DIR%\client.exe
)

echo.
echo ✓ Build completed successfully!
echo Version: %VERSION%
echo Output directory: %OUTPUT_DIR%

:eof
