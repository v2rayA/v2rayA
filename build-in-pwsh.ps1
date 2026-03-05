Write-Host $PSScriptRoot

Function Get-build-tools(){
    if ([String]::IsNullOrEmpty($(Get-Command git -ErrorAction ignore))) {
        Write-Output "You don't install git, please install it and add it to your path."
        $We_Should_Exit_Now = $true
    }
    if ([String]::IsNullOrEmpty($(Get-Command yarn -ErrorAction ignore))) {
        Write-Output "You don't install yarn, please install it and add it to your path."
        Write-Output "You should also install Node.js to make yarn work fine."
        $We_Should_Exit_Now = $true
    }
    if ([String]::IsNullOrEmpty($(Get-Command go -ErrorAction ignore))) {
        Write-Output "You don't install go, please install it and add it to your path."
        $We_Should_Exit_Now = $true
    }
    if ($We_Should_Exit_Now) {
        Exit 1
    }
}

Function Build-v2rayA(){
    #Get OS
    if ($env:GOOS -eq "windows") {
        $v2rayaBin = "v2raya.exe"
    } elseif ($env:GOOS -ne "windows") {
         $v2rayaBin = "v2raya"
    } elseif ($env:WinDir) {
        $v2rayaBin = "v2raya.exe"
    } else {
        $v2rayaBin = "v2raya"
    }
    #Get Paths
    $TerminalPath = Get-Item -LiteralPath ./ | ForEach-Object  -Process { $_.FullName }
    $CurrentPath = $PSScriptRoot
    Set-Location -Path "$CurrentPath"
    #Get Informations
    $DateLong = git log -1 --format="%cd" --date=short
    $Date = $DateLong -replace "-"; ""
    $count = git rev-list --count HEAD
    $commit = git rev-parse --short HEAD
    #Version
    $version = "unstable-$date.r$count.$commit"
    #Disable CGO
    ${env:CGO_ENABLED} = "0"
    #Set yarn's output path
    ${env:OUTPUT_DIR} = "$CurrentPath/service/server/router/web"
    #Build Web Panel
    Set-Location -Path "$CurrentPath/gui"
    yarn; yarn build
    #Build v2rayA
    Set-Location -Path "$CurrentPath/service"
    go build -tags "with_gvisor" -ldflags "-X github.com/v2rayA/v2rayA/conf.Version=$version -s -w" -o "$CurrentPath/$v2rayaBin"
    Set-Location -Path "$TerminalPath"
}

Set-PSDebug -Trace 1

Get-build-tools
Build-v2rayA

Set-PSDebug -Trace 0