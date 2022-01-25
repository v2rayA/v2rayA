Function Gzip-File([ValidateScript({Test-Path $_})][string]$File){
 
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

Set-PSDebug -Trace 1

## Check OS
$TestWinDir = Test-Path $env:windir -ErrorAction Ignore

if ([String]::IsNullOrEmpty($TestWinDir)) { 
    $v2rayaBin = "v2raya"
}
else {
    $v2rayaBin = "v2raya.exe"
}


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
            ## Get current folder
            ## $SHELL_FOLDER = Get-Item -LiteralPath ./ | ForEach-Object  -Process { $_.FullName }
            $CWD = Get-Location
            $shell_path = Resolve-Path -Path $PSCommandPath
            $SHELL_FOLDER = Split-Path $shell_path

            ## Get date
            $DateLong = git log -1 --format="%cd" --date=short
            $Date = $DateLong -replace "-"; ""

            ## Other info
            $count = git rev-list --count HEAD
            $commit = git rev-parse --short HEAD

            ## Version
            $version = "unstable-$date.r$count.$commit"

            ## Disable CGO
            ${env:CGO_ENABLED} = "0"

            ## Set yarn's output path
            ${env:OUTPUT_DIR} = "$SHELL_FOLDER/service/server/router/web"

            ## Build v2rayA
            $guiPath = $SHELL_FOLDER + "/gui"
            Set-Location -path "$guiPath"
            yarn
            yarn build
            Get-ChildItem "$SHELL_FOLDER/service/server/router/web" -recurse |Where-Object{$_.PSIsContainer -eq $False}|ForEach-Object -Process{
                if($_.Extension -ne ".png" -and $_.Extension -ne ".gz" -and $_.Name -ne "index.html"){
                    Gzip-File($_.FullName)
                    Remove-Item -Path $_.FullName
                }
            }
            $corePath = $SHELL_FOLDER + "/service"
            Set-Location -path "$corePath"
            go build -ldflags "-X github.com/v2rayA/v2rayA/conf.Version=$version -s -w" -o "$SHELL_FOLDER/$v2rayaBin"
        }
    }
}

Set-Location -Path "$SHELL_FOLDER"