## Check OS
$TestWinDir = Test-Path $env:windir -ErrorAction Ignore

if ([String]::IsNullOrEmpty($TestWinDir)) { 
    $v2rayaBin = "v2raya"
}
else {
    $v2rayaBin = "v2raya.exe"
}

## Get current folder
$SHELL_FOLDER = Get-Item -LiteralPath .\ | ForEach-Object  -Process { $_.FullName }

## Get date
$DateLong = git log -1 --format="%cd" --date=short
$Date = $DateLong -replace "-"; ""

## Other info
$count = git rev-list --count HEAD
$commit = git rev-parse --short HEAD

## Version
$version = "unstable-$date.r$count.$commit"

## Test Git / yarn /Go
$TestGit = Get-Command git -ErrorAction ignore
$TestYarn = Get-Command yarn -ErrorAction ignore
$TestGo = Get-Command go -ErrorAction ignore

if ([String]::IsNullOrEmpty($TestGit)) {
    Write-Output "You don't install git, please install it and add it to your path."
}
else {
    if ([String]::IsNullOrEmpty($TestYarn)) {
        Write-Output "You don't install yarn, please install it and add it to your path."
    }
    else {
        if ([String]::IsNullOrEmpty($TestGo)) {
            Write-Output "You don't install golang, please install it and add it to your path."
        }
        else {
            ${env:CGO_ENABLED} = "0"
            ${env:OUTPUT_DIR} = "$SHELL_FOLDER/service/server/router/web"
            $guiPath = $SHELL_FOLDER + "/gui"
            Set-Location -path "$guiPath"
            yarn
            yarn build
            $corePath = $SHELL_FOLDER + "/service"
            Set-Location -path "$corePath"
            go build -ldflags "-X github.com/v2rayA/v2rayA/conf.Version=$version -s -w" -o "$SHELL_FOLDER/$v2rayaBin"
        }
    }
}
