Write-Host $PSScriptRoot

Function Compress-File([ValidateScript({Test-Path $_})][string]$File){
 
    $srcFile = Get-Item -Path $File
    $newFileName = "$($srcFile.FullName).gz"
 
    try
    {
        $srcFileStream = New-Object System.IO.FileStream($srcFile.FullName,([IO.FileMode]::Open),([IO.FileAccess]::Read),([IO.FileShare]::Read))
        $dstFileStream = New-Object System.IO.FileStream($newFileName,([IO.FileMode]::Create),([IO.FileAccess]::Write),([IO.FileShare]::None))
        $gzip = New-Object System.IO.Compression.GZipStream($dstFileStream,[System.IO.Compression.CompressionMode]::Compress)
        $srcFileStream.CopyTo($gzip)
    } 
    catch
    {
        Write-Host "$_.Exception.Message" -ForegroundColor Red
    }
    finally
    {
        $gzip.Dispose()
        $srcFileStream.Dispose()
        $dstFileStream.Dispose()
    }
}

Function Get-build-tools(){
    if ([String]::IsNullOrEmpty($(Get-Command git -ErrorAction ignore))) {
        Write-Output "You don't install git, please install it and add it to your path."
    }
    if ([String]::IsNullOrEmpty($(Get-Command yarn -ErrorAction ignore))) {
        Write-Output "You don't install yarn, please install it and add it to your path."
        Write-Output "You should also install Node.js to make yarn work fine."
    }
    if ([String]::IsNullOrEmpty($(Get-Command go -ErrorAction ignore))) {
        Write-Output "You don't install go, please install it and add it to your path."
    }
    if ([String]::IsNullOrEmpty($(Get-Command git -ErrorAction ignore))) {
        exit 1
    }
    if ([String]::IsNullOrEmpty($(Get-Command yarn -ErrorAction ignore))) {
        exit 1
    }
    if ([String]::IsNullOrEmpty($(Get-Command go -ErrorAction ignore))) {
        exit 1
    }
}

Function Build-v2rayA(){
    #Get OS
    if ([String]::IsNullOrEmpty($(Test-Path ${env:windir} -ErrorAction Ignore))) { 
        $v2rayaBin = "v2raya"
    }
    else {
        $v2rayaBin = "v2raya.exe"
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
    #Compress Web Panel's files
    Get-ChildItem "$CurrentPath/service/server/router/web" -recurse |Where-Object{$_.PSIsContainer -eq $False}|ForEach-Object -Process{
        if($_.Extension -ne ".png" -and $_.Extension -ne ".gz" -and $_.Name -ne "index.html"){
            Compress-File($_.FullName)
            Remove-Item -Path $_.FullName
        }
    }
    #Build v2rayA
    Set-Location -Path "$CurrentPath/service"
    go build -ldflags "-X github.com/v2rayA/v2rayA/conf.Version=$version -s -w" -o "$CurrentPath/$v2rayaBin"
    Set-Location -Path "$TerminalPath"
}

Set-PSDebug -Trace 1

Get-build-tools
Build-v2rayA

Set-PSDebug -Trace 0