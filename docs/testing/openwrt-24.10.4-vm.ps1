param(
    [Parameter(Mandatory = $true)] [string] $Qemu,
    [Parameter(Mandatory = $true)] [string] $Go,
    [string] $WorkDir = (Join-Path ([IO.Path]::GetTempPath()) "v2raya-openwrt-24.10.4"),
    [int] $SshPort = 25922,
    [int] $SerialPort = 25923,
    [switch] $KeepVM
)

$ErrorActionPreference = "Stop"
$nullDevice = if ($IsWindows) { "NUL" } else { "/dev/null" }
$release = "24.10.4"
$imageName = "openwrt-$release-x86-64-generic-ext4-combined.img.gz"
$imageUrl = "https://downloads.openwrt.org/releases/$release/targets/x86/64/$imageName"
$imageSha256 = "57f43d2a665646b0c45b909328704796db1777f85cc0c8ad51ec9b830d34ffc5"
$repository = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
$Qemu = (Resolve-Path -LiteralPath $Qemu).Path
$Go = (Resolve-Path -LiteralPath $Go).Path
New-Item -ItemType Directory -Force -Path $WorkDir | Out-Null
$WorkDir = (Resolve-Path -LiteralPath $WorkDir).Path
$service = Join-Path $repository "service"
$compressed = Join-Path $WorkDir $imageName
$baseImage = Join-Path $WorkDir "openwrt-$release.img"
$runImage = Join-Path $WorkDir "openwrt-$release-run.img"
$serviceTest = Join-Path $WorkDir "server-service.test"
$kernelTest = Join-Path $WorkDir "kernel-v2ray.test"
$sshKey = Join-Path $WorkDir "id_ed25519"
$serialLog = Join-Path $WorkDir "serial.log"
$qemuErrorLog = Join-Path $WorkDir "qemu-error.log"

function Invoke-Native([string] $File, [string[]] $Arguments) {
    & $File @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "$File exited with $LASTEXITCODE"
    }
}

function Invoke-Ssh([string] $Command, [switch] $Quiet) {
    $arguments = @(
        "-F", $nullDevice,
        "-p", "$SshPort",
        "-i", $sshKey,
        "-o", "BatchMode=yes",
        "-o", "IdentitiesOnly=yes",
        "-o", "ConnectTimeout=3",
        "-o", "StrictHostKeyChecking=no",
        "-o", "UserKnownHostsFile=$nullDevice",
        "root@127.0.0.1", $Command
    )
    if ($Quiet) {
        $script:lastSshError = (& ssh @arguments 2>&1 | Out-String).Trim()
        return $LASTEXITCODE -eq 0
    }
    Invoke-Native "ssh" $arguments
}

if (-not (Test-Path -LiteralPath $compressed)) {
    Invoke-WebRequest -Uri $imageUrl -OutFile $compressed
}
$actualHash = (Get-FileHash -LiteralPath $compressed -Algorithm SHA256).Hash.ToLowerInvariant()
if ($actualHash -ne $imageSha256) {
    throw "OpenWrt image SHA-256 is $actualHash, expected $imageSha256"
}
if (-not (Test-Path -LiteralPath $baseImage)) {
    $source = [IO.File]::OpenRead($compressed)
    try {
        $gzip = [IO.Compression.GZipStream]::new($source, [IO.Compression.CompressionMode]::Decompress)
        try {
            $target = [IO.File]::Create($baseImage)
            try { $gzip.CopyTo($target) } finally { $target.Dispose() }
        } finally { $gzip.Dispose() }
    } finally { $source.Dispose() }
}
Copy-Item -LiteralPath $baseImage -Destination $runImage -Force

if (-not (Test-Path -LiteralPath $sshKey)) {
    Invoke-Native "ssh-keygen" @("-q", "-t", "ed25519", "-N", "", "-f", $sshKey)
}
$publicKeyParts = (Get-Content -LiteralPath "$sshKey.pub" -Raw).Trim() -split "\s+"
$sshPublicKey = $publicKeyParts[0..1] -join " "

$oldGoos, $oldGoarch, $oldCgo = $env:GOOS, $env:GOARCH, $env:CGO_ENABLED
$env:GOOS, $env:GOARCH, $env:CGO_ENABLED = "linux", "amd64", "0"
$env:GOCACHE = Join-Path $WorkDir "go-build"
try {
    Push-Location $service
    try {
        Invoke-Native $Go @("test", "-c", "-o", $serviceTest, "./server/service")
        Invoke-Native $Go @("test", "-c", "-o", $kernelTest, "./kernel/v2ray")
    } finally { Pop-Location }
} finally {
    $env:GOOS, $env:GOARCH, $env:CGO_ENABLED = $oldGoos, $oldGoarch, $oldCgo
}

$qemuArguments = @(
    "-machine", "q35,accel=tcg",
    "-cpu", "max",
    "-smp", "2",
    "-m", "256",
    "-display", "none",
    "-no-reboot",
    "-serial", "tcp:127.0.0.1:${SerialPort},server=on,wait=off",
    "-drive", "file=$runImage,format=raw,if=virtio",
    "-netdev", "user,id=net0,hostfwd=tcp:127.0.0.1:${SshPort}-:22",
    "-device", "virtio-net-pci,netdev=net0"
)
$startInfo = [Diagnostics.ProcessStartInfo]::new()
$startInfo.FileName = $Qemu
$startInfo.UseShellExecute = $false
$startInfo.CreateNoWindow = $true
$startInfo.RedirectStandardOutput = $true
$startInfo.RedirectStandardError = $true
foreach ($argument in $qemuArguments) { [void] $startInfo.ArgumentList.Add($argument) }
$qemuProcess = [Diagnostics.Process]::new()
$qemuProcess.StartInfo = $startInfo
if (-not $qemuProcess.Start()) { throw "QEMU did not start" }
$stdout = $qemuProcess.StandardOutput.ReadToEndAsync()
$stderr = $qemuProcess.StandardError.ReadToEndAsync()
$serialClient = $null
$serialStream = $null
$serialText = [Text.StringBuilder]::new()

function Wait-Serial([string] $Pattern, [DateTime] $Deadline) {
    $buffer = [byte[]]::new(4096)
    while ([DateTime]::UtcNow -lt $Deadline -and -not $qemuProcess.HasExited) {
        if ($serialStream.DataAvailable) {
            $count = $serialStream.Read($buffer, 0, $buffer.Length)
            if ($count -gt 0) {
                [void] $serialText.Append([Text.Encoding]::UTF8.GetString($buffer, 0, $count))
                if ($serialText.ToString().Contains($Pattern)) { return }
            }
        } else {
            Start-Sleep -Milliseconds 100
        }
    }
    throw "Serial console did not reach marker: $Pattern"
}

try {
    $serialDeadline = [DateTime]::UtcNow.AddMinutes(4)
    while ([DateTime]::UtcNow -lt $serialDeadline -and $null -eq $serialClient) {
        try {
            $candidate = [Net.Sockets.TcpClient]::new()
            $candidate.Connect("127.0.0.1", $SerialPort)
            $serialClient = $candidate
            $serialStream = $serialClient.GetStream()
        } catch {
            if ($null -ne $candidate) { $candidate.Dispose() }
            Start-Sleep -Milliseconds 200
        }
    }
    if ($null -eq $serialClient) { throw "Could not connect to QEMU serial console" }
    Wait-Serial "Please press Enter to activate this console." $serialDeadline
    $enter = [Text.Encoding]::ASCII.GetBytes("`r`n")
    $serialStream.Write($enter, 0, $enter.Length)
    # The stock image has no hostname during early boot, so the prompt is
    # usually root@(none):~#. Match the stable prompt prefix instead.
    Wait-Serial "root@" $serialDeadline
    # The serial login prompt appears several seconds before netifd creates
    # br-lan and starts dropbear. Wait for the bridge to be forwarding.
    Wait-Serial "br-lan: port 1(eth0) entered forwarding state" $serialDeadline
    # The stock x86 image assigns its only NIC to the static LAN bridge. Add
    # QEMU's conventional guest address without changing the base image.
    $networkCommand = "ip addr add 10.0.2.15/24 dev br-lan 2>/dev/null || true; ip link set br-lan up; mkdir -p /etc/dropbear; printf '%s\n' '$sshPublicKey' > /etc/dropbear/authorized_keys; chmod 600 /etc/dropbear/authorized_keys; /etc/init.d/dropbear restart; printf '__V2RAYA_%s__\n' NETWORK_READY`r`n"
    $networkBytes = [Text.Encoding]::ASCII.GetBytes($networkCommand)
    $serialStream.Write($networkBytes, 0, $networkBytes.Length)
    Wait-Serial "__V2RAYA_NETWORK_READY__" $serialDeadline

    $deadline = [DateTime]::UtcNow.AddMinutes(1)
    $ready = $false
    while ([DateTime]::UtcNow -lt $deadline -and -not $qemuProcess.HasExited) {
        if (Invoke-Ssh "true" -Quiet) { $ready = $true; break }
        Start-Sleep -Seconds 2
    }
    if (-not $ready) {
        if ($lastSshError) { Write-Host $lastSshError }
        throw "OpenWrt SSH did not become ready"
    }

    $verifyCommand = '. /etc/openwrt_release; test "$DISTRIB_RELEASE" = ''{0}''; test "$(uname -m)" = x86_64; echo OpenWrt-$DISTRIB_RELEASE-$(uname -m)' -f $release
    Invoke-Ssh $verifyCommand

    $scpCommon = @(
        "-O",
        "-F", $nullDevice,
        "-P", "$SshPort",
        "-i", $sshKey,
        "-o", "BatchMode=yes",
        "-o", "IdentitiesOnly=yes",
        "-o", "StrictHostKeyChecking=no",
        "-o", "UserKnownHostsFile=$nullDevice"
    )
    Invoke-Native "scp" ($scpCommon + @($serviceTest, "root@127.0.0.1:/tmp/server-service.test"))
    Invoke-Ssh "chmod +x /tmp/server-service.test && /tmp/server-service.test -test.v -test.run 'Test(RandomStrategyUsesLowestNonEmptyLatencyBucket|FirstAvailableUsesStableGroupOrder|KeepCurrentChangesOnlyAfterFailureAndFailsClosed|ForcedAutomaticGroupRefreshIgnoresFutureSchedule)$' && rm -f /tmp/server-service.test"

    Invoke-Native "scp" ($scpCommon + @($kernelTest, "root@127.0.0.1:/tmp/kernel-v2ray.test"))
    Invoke-Ssh "chmod +x /tmp/kernel-v2ray.test && /tmp/kernel-v2ray.test -test.v -test.run 'Test(ApplySelectionUsesWorkerChoiceForRandomAndFirstAvailable|EmptyManagedProxyBlocksOnlyItsOwnTraffic|NativeGroupStrategiesFailClosed)$' && rm -f /tmp/kernel-v2ray.test"
} finally {
    if (-not $qemuProcess.HasExited) {
        $qemuProcess.Kill($true)
        $qemuProcess.WaitForExit()
    }
    if ($null -ne $serialStream) {
        $buffer = [byte[]]::new(4096)
        while ($serialStream.DataAvailable) {
            $count = $serialStream.Read($buffer, 0, $buffer.Length)
            if ($count -le 0) { break }
            [void] $serialText.Append([Text.Encoding]::UTF8.GetString($buffer, 0, $count))
        }
    }
    [IO.File]::WriteAllText($serialLog, $serialText.ToString())
    [IO.File]::WriteAllText($qemuErrorLog, $stdout.GetAwaiter().GetResult() + $stderr.GetAwaiter().GetResult())
    if ($null -ne $serialStream) { $serialStream.Dispose() }
    if ($null -ne $serialClient) { $serialClient.Dispose() }
    $qemuProcess.Dispose()
    if (-not $KeepVM) { Remove-Item -LiteralPath $runImage -Force -ErrorAction SilentlyContinue }
}

Write-Host "OpenWrt $release VM checks passed; serial log: $serialLog"
