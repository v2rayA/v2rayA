$ErrorActionPreference = "Stop"

$rootDir = $PSScriptRoot
if ($args.Count -gt 0) {
    $version = $args[0]
} else {
    $version = (& git -C $rootDir describe --tags --always --dirty).Trim()
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}
$webDir = Join-Path $rootDir "service/server/router/web"

& yarn --cwd (Join-Path $rootDir "gui") --check-files
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
& yarn --cwd (Join-Path $rootDir "gui") build
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

Remove-Item -Recurse -Force -ErrorAction Ignore $webDir
New-Item -ItemType Directory -Force $webDir | Out-Null
Get-ChildItem -Force (Join-Path $rootDir "web") | Copy-Item -Destination $webDir -Recurse -Force

Push-Location (Join-Path $rootDir "core")
try {
    $env:CGO_ENABLED = "0"
    & go build -trimpath -o (Join-Path $rootDir "v2raya_core.exe") "-ldflags=-X main.Version=$version -s -w" ./main
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
} finally {
    Pop-Location
}

Push-Location (Join-Path $rootDir "service")
try {
    $env:CGO_ENABLED = "0"
    & go build -trimpath -o (Join-Path $rootDir "v2raya.exe") "-ldflags=-X github.com/v2rayA/v2rayA/conf.Version=$version -s -w"
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
} finally {
    Pop-Location
}
