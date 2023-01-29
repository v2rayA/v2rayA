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

Invoke-Expression "& {$(Invoke-RestMethod get.scoop.sh)} -RunAsAdmin"
# scoop bucket add versions;scoop install nodejs16;
scoop install yarn go nodejs

${env:NODE_OPTIONS} = "--openssl-legacy-provider"

yarn --cwd gui --check-files
yarn --cwd gui build

Get-ChildItem "./web" -recurse |Where-Object{$_.PSIsContainer -eq $False}|ForEach-Object -Process{
    if($_.Extension -ne ".png" -and $_.Extension -ne ".gz" -and $_.Name -ne "index.html"){
        Compress-File($_.FullName)
        Remove-Item -Path $_.FullName
    }
}

Copy-Item -Path ./web ./service/server/router/ -Recurse

New-Item -ItemType Directory -Path ./ -Name "v2raya-x86_64-windows"; New-Item -ItemType Directory -Path ".\v2raya-x86_64-windows\bin"
New-Item -ItemType Directory -Path ./ -Name "v2raya-arm64-windows"; New-Item -ItemType Directory -Path ".\v2raya-arm64-windows\bin"

Set-Location -Path ./service
$VERSION = ${env:VERSION}
$env:CGO_ENABLED = "0"
$build_flags = "-X github.com/v2rayA/v2rayA/conf.Version=$VERSION -s -w"
$env:GOARCH = "amd64"; $env:GOOS = "windows"; go build -ldflags $build_flags -o '../v2raya-x86_64-windows/bin/v2raya.exe'
$env:GOARCH = "arm64"; $env:GOOS = "windows"; go build -ldflags $build_flags -o '../v2raya-arm64-windows/bin/v2raya.exe'

Set-Location ../

Copy-Item "./install/windows-inno/v2raya.ico" "D:\v2raya.ico"

Copy-Item "./v2raya-x86_64-windows/" "D:\" -Recurse
New-Item -ItemType Directory -Path "D:\v2raya-x86_64-windows\data"

Copy-Item "./v2raya-arm64-windows/" "D:\" -Recurse
New-Item -ItemType Directory -Path "D:\v2raya-arm64-windows\data"

$Version_v2ray = ((Invoke-WebRequest -Uri 'https://api.github.com/repos/v2fly/v2ray-core/releases/latest' | ConvertFrom-Json).tag_name).Split("v")[1]
$Url_v2ray_x64 = "https://github.com/v2fly/v2ray-core/releases/download/v$Version_v2ray/v2ray-windows-64.zip"
$Url_v2ray_A64 = "https://github.com/v2fly/v2ray-core/releases/download/v$Version_v2ray/v2ray-windows-arm64-v8a.zip"

Invoke-WebRequest $Url_v2ray_x64 -OutFile "D:\v2ray-windows-x64.zip"
Expand-Archive -Path "D:\v2ray-windows-x64.zip" -DestinationPath "D:\v2raya-x86_64-windows\bin\"
Move-Item -Path "D:\v2raya-x86_64-windows\bin\*.dat" -Destination "D:\v2raya-x86_64-windows\data"
Remove-Item -Path "D:\v2raya-x86_64-windows\bin\wv2ray.exe" -Force -Recurse -ErrorAction SilentlyContinue
Remove-Item -Path "D:\v2raya-x86_64-windows\bin\*.json" -Force -Recurse -ErrorAction SilentlyContinue

Invoke-WebRequest $Url_v2ray_A64 -OutFile "D:\v2ray-windows-A64.zip"
Expand-Archive -Path "D:\v2ray-windows-A64.zip" -DestinationPath "D:\v2raya-arm64-windows\bin\"
Move-Item -Path "D:\v2raya-arm64-windows\bin\*.dat" -Destination "D:\v2raya-arm64-windows\data"
Remove-Item -Path "D:\v2raya-arm64-windows\bin\wv2ray.exe" -Force -Recurse -ErrorAction SilentlyContinue
Remove-Item -Path "D:\v2raya-arm64-windows\bin\*.json" -Force -Recurse -ErrorAction SilentlyContinue

$Url_WinSW = "https://github.com/winsw/winsw/releases/download/v3.0.0-alpha.10/WinSW-net461.exe"
Invoke-WebRequest $Url_WinSW -OutFile "D:\WinSW.exe"
Copy-Item -Path "D:\WinSW.exe" -Destination "D:\v2raya-x86_64-windows\v2rayA-service.exe"
Copy-Item -Path "D:\WinSW.exe" -Destination "D:\v2raya-arm64-windows\v2rayA-service.exe"

Set-Content -Value '<!--
MIT License
Copyright (c) 2008-2020 Kohsuke Kawaguchi, Sun Microsystems, Inc., CloudBees,
Inc., Oleg Nenashev and other contributors
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
-->

<service>
<id>v2rayA</id>
<name>v2rayA background service for Windows</name>
<description>v2rayA is a V2Ray client, compatible with SS, SSR, Trojan(trojan-go), PingTunnel protocols.</description>
<executable>%BASE%\bin\v2raya.exe</executable>
<arguments>--lite --log-file "v2raya.log" --v2ray-bin "%BASE%\bin\v2ray.exe" --v2ray-assetsdir "%BASE%\data" --config "%BASE%"</arguments>
<workingdirectory>%TEMP%</workingdirectory>
<log mode="roll"></log>
<onfailure action="restart" delay="10 sec"/>
</service>
' -Path "D:\v2raya-x86_64-windows\v2rayA-service.xml"
Copy-Item -Path "D:\v2raya-x86_64-windows\v2rayA-service.xml" -Destination "D:\v2raya-arm64-windows\v2rayA-service.xml"

Invoke-WebRequest -Uri "https://raw.githubusercontent.com/v2rayA/v2rayA/feat_v5/LICENSE" -OutFile "D:\LICENSE.txt"

$(Get-Content -Path .\install\windows-inno\windows_x86_64.iss).replace("TheRealVersion", "$VERSION") | Out-File "D:\windows_x86_64.iss"
$(Get-Content -Path .\install\windows-inno\windows_arm64.iss).replace("TheRealVersion", "$VERSION") | Out-File "D:\windows_arm64.iss"

& 'C:\Program Files (x86)\Inno Setup 6\ISCC.exe' "D:\windows_x86_64.iss"
& 'C:\Program Files (x86)\Inno Setup 6\ISCC.exe' "D:\windows_arm64.iss"

Copy-Item "D:\installer_windows_inno_x64.exe"  ".\installer_windows_inno_x64.exe"
Copy-Item "D:\installer_windows_inno_arm64.exe"  ".\installer_windows_inno_arm64.exe"
Copy-Item "./v2raya-x86_64-windows/bin/v2raya.exe" "./v2raya_windows_x64_$VERSION.exe"
Copy-Item "./v2raya-arm64-windows/bin/v2raya.exe" "./v2raya_windows_arm64_$VERSION.exe"

foreach ($file in Get-ChildItem -Path .\ -Filter "*.exe" -Recurse) {
    (Get-FileHash $file).Hash | Out-File "$file"'.sha256.txt'
}