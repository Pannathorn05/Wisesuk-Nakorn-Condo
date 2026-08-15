<#
.SYNOPSIS
    อัปโหลดรูปสาขาจาก backend/uploads/<slug>/ เข้าฐานข้อมูล แล้วตรวจว่าเข้าครบจริง

.DESCRIPTION
    รูปถูกเก็บเป็นเนื้อไฟล์ในตาราง assets ไม่ใช่ไฟล์บนดิสก์ ทุกครั้งที่ล้าง volume
    ของฐานข้อมูล (docker compose down -v) รูปจะหายไปด้วย ให้รันสคริปต์นี้ลงใหม่

    โครงสร้างโฟลเดอร์ที่สคริปต์อ่าน — ชื่อโฟลเดอร์ชั้นแรกต้องตรงกับ slug ของสาขา:

        uploads/prachauthit-45/day/*.jpg
        uploads/prachauthit-45/month/*.jpg
        uploads/bangkhae/*.jpg
        uploads/charoenkrung-place/*.jpg

    ถ้าโฟลเดอร์ย่อยมีไฟล์ caption.txt (UTF-8) ข้อความในไฟล์จะถูกใช้เป็นคำบรรยายของ
    ทุกรูปในโฟลเดอร์นั้น — วิธีนี้เลี่ยงปัญหา PowerShell 5.1 ส่ง argument ภาษาไทย
    เป็น CP874 แล้วโดน API ปฏิเสธด้วย 422

    ท้ายสคริปต์จะเทียบจำนวนไฟล์กับจำนวนรูปที่ API ตอบกลับ ถ้าไม่ตรงจะจบด้วย exit code 1

.PARAMETER Replace
    ลบรูปเดิมของสาขานั้นทิ้งก่อนแล้วลงใหม่ ถ้าไม่ระบุ สาขาที่มีรูปอยู่แล้วจะถูกข้าม

.EXAMPLE
    .\scripts\upload-branch-images.ps1

.EXAMPLE
    .\scripts\upload-branch-images.ps1 -Replace
#>
[CmdletBinding()]
param(
    [string] $BaseUrl   = "http://localhost:8080",
    [string] $Password  = "Wisetsuk!2026",
    [string] $UploadDir = (Join-Path $PSScriptRoot "..\uploads"),
    [switch] $Replace
)

$ErrorActionPreference = "Stop"

# slug ของสาขา -> บัญชีแอดมินที่ดูแลสาขานั้น
# ใช้บัญชี admin ไม่ใช่ superadmin เพราะ token ของ admin ผูกสาขาไว้แล้ว ไม่ต้องส่ง branch_id
$adminOf = @{
    "prachauthit-45"     = "admin.1@wisetsuk.com"
    "bangkhae"           = "admin.2@wisetsuk.com"
    "charoenkrung-place" = "admin.3@wisetsuk.com"
}

function Get-Token {
    param([string] $Email)

    $body = @{ email = $Email; password = $Password } | ConvertTo-Json -Compress
    $res  = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/v1/auth/login" `
                              -ContentType "application/json" -Body $body
    return $res.data.access_token
}

function Get-Branch {
    param([string] $Slug)

    $all = (Invoke-RestMethod -Uri "$BaseUrl/api/v1/branches").data
    return $all | Where-Object { $_.slug -eq $Slug }
}

# Get-BranchImages คืนรูปของสาขาเป็น array เสมอ
#
# API ตัดฟิลด์ images ออกจาก JSON เมื่อสาขายังไม่มีรูป (omitempty) ทำให้ค่าเป็น $null
# และ @($null).Count ใน PowerShell คืน 1 ไม่ใช่ 0 — ต้องกรอง $null ทิ้งก่อน
# ไม่งั้นสาขาที่ไม่มีรูปเลยจะถูกนับว่ามี 1 รูป
function Get-BranchImages {
    param([string] $BranchId)

    $b = (Invoke-RestMethod -Uri "$BaseUrl/api/v1/branches/$BranchId").data
    return @($b.images | Where-Object { $null -ne $_ })
}

function Remove-BranchImages {
    param([string] $BranchId, [string] $Token)

    foreach ($img in (Get-BranchImages -BranchId $BranchId)) {
        Invoke-RestMethod -Method Delete -Headers @{ Authorization = "Bearer $Token" } `
            -Uri "$BaseUrl/api/v1/admin/branch/images/$($img.id)" | Out-Null
    }
}

# ---------------------------------------------------------------- เริ่มทำงาน

if (-not (Test-Path $UploadDir)) {
    throw "ไม่พบโฟลเดอร์รูป: $UploadDir"
}

Write-Host "อ่านรูปจาก $((Resolve-Path $UploadDir).Path)`n"

$totalUploaded = 0
$problems      = @()

foreach ($slug in $adminOf.Keys | Sort-Object) {
    $dir = Join-Path $UploadDir $slug
    if (-not (Test-Path $dir)) {
        Write-Host "ข้าม $slug — ไม่มีโฟลเดอร์"
        continue
    }

    $files = @(Get-ChildItem -Path $dir -Recurse -File -Include *.jpg, *.jpeg, *.png, *.webp |
               Sort-Object FullName)
    if ($files.Count -eq 0) {
        Write-Host "ข้าม $slug — ไม่มีไฟล์รูป"
        continue
    }

    $branch = Get-Branch -Slug $slug
    if ($null -eq $branch) {
        $problems += "ไม่พบสาขา $slug ในฐานข้อมูล (รัน seed หรือยัง?)"
        continue
    }

    $token   = Get-Token -Email $adminOf[$slug]
    $already = (Get-BranchImages -BranchId $branch.id).Count

    if ($already -gt 0 -and -not $Replace) {
        Write-Host "ข้าม $slug — มีรูปอยู่แล้ว $already รูป (ใช้ -Replace เพื่อลงใหม่ทับ)"
        continue
    }
    if ($already -gt 0) {
        Write-Host "ลบรูปเดิมของ $slug ทิ้ง $already รูป"
        Remove-BranchImages -BranchId $branch.id -Token $token
    }

    Write-Host "$slug ($($branch.name)) — $($files.Count) ไฟล์"

    $i = 0
    foreach ($f in $files) {
        $curlArgs = @(
            "-s", "-o", "NUL", "-w", "%{http_code}",
            "-X", "POST", "$BaseUrl/api/v1/admin/branch/images/upload",
            "-H", "Authorization: Bearer $token",
            "-F", "image=@$($f.FullName)",
            "-F", "sort_order=$i"
        )

        # คำบรรยายมาจากไฟล์ caption.txt ในโฟลเดอร์เดียวกับรูป (ถ้ามี)
        $capFile = Join-Path $f.DirectoryName "caption.txt"
        if (Test-Path $capFile) {
            $curlArgs += @("-F", "caption=<$capFile")
        }

        $code = & curl.exe @curlArgs
        if ($code -eq "201") {
            $totalUploaded++
            Write-Host ("  [{0,3}] {1}" -f $code, $f.Name)
        } else {
            $problems += "$slug/$($f.Name) -> HTTP $code"
            Write-Host ("  [{0,3}] {1}  <-- ไม่สำเร็จ" -f $code, $f.Name) -ForegroundColor Red
        }
        $i++
    }
    Write-Host ""
}

# ---------------------------------------------------------------- ตรวจผล

Write-Host "ตรวจผลจาก API"

foreach ($slug in $adminOf.Keys | Sort-Object) {
    $dir = Join-Path $UploadDir $slug
    if (-not (Test-Path $dir)) { continue }

    $expected = @(Get-ChildItem -Path $dir -Recurse -File -Include *.jpg, *.jpeg, *.png, *.webp).Count
    $branch   = Get-Branch -Slug $slug
    if ($null -eq $branch) { continue }

    $actual = (Get-BranchImages -BranchId $branch.id).Count
    $mark   = "ตรง"
    if ($expected -ne $actual) {
        $mark = "ไม่ตรง"
        $problems += "$slug — ไฟล์ $expected แต่ใน API มี $actual"
    }
    Write-Host ("  {0,-20} ไฟล์ {1,2}  ใน API {2,2}  {3}" -f $slug, $expected, $actual, $mark)
}

Write-Host ""
if ($problems.Count -gt 0) {
    Write-Host "มีปัญหา $($problems.Count) รายการ:" -ForegroundColor Red
    $problems | ForEach-Object { Write-Host "  - $_" -ForegroundColor Red }
    exit 1
}

Write-Host "อัปโหลดสำเร็จ $totalUploaded รูป ทุกสาขาตรงกับจำนวนไฟล์" -ForegroundColor Green
