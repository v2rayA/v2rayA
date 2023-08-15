Copy-Item "./install/windows-inno/v2raya.ico" "D:\v2raya.ico"

$Version_v2ray = ((Invoke-RestMethod -Uri 'https://api.github.com/repos/v2fly/v2ray-core/releases/latest').tag_name).Split("v")[1]
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

$Url_WinSW = "https://github.com/winsw/winsw/releases/download/v3.0.0-alpha.11/WinSW-net461.exe"
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
<description>v2rayA is a V2Ray client, compatible with SS, SSR, Trojan(trojan-go), Juicity protocols.</description>
<executable>%BASE%\bin\v2raya.exe</executable>
<env name="PATH" value="%BASE%\bin\;%windir%\system32\"/>
<arguments>--lite --log-file "v2raya.log" --v2ray-assetsdir "%BASE%\data" --config "%BASE%"</arguments>
<workingdirectory>%TEMP%</workingdirectory>
<log mode="roll"></log>
<onfailure action="restart" delay="10 sec"/>
</service>
' -Path "D:\v2raya-x86_64-windows\v2rayA-service.xml"
Copy-Item -Path "D:\v2raya-x86_64-windows\v2rayA-service.xml" -Destination "D:\v2raya-arm64-windows\v2rayA-service.xml"
Copy-Item -Path ".\LICENSE" "D:\LICENSE.txt"

$(Get-Content -Path .\install\windows-inno\windows_x86_64.iss).replace("TheRealVersion", "$VERSION") | Out-File "D:\windows_x86_64.iss"
$(Get-Content -Path .\install\windows-inno\windows_arm64.iss).replace("TheRealVersion", "$VERSION") | Out-File "D:\windows_arm64.iss"

& 'C:\Program Files (x86)\Inno Setup 6\ISCC.exe' "D:\windows_x86_64.iss"
& 'C:\Program Files (x86)\Inno Setup 6\ISCC.exe' "D:\windows_arm64.iss"