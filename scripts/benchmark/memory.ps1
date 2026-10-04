<#
.SYNOPSIS
  Benchmark de memória do Sortly no Windows: Electron 1.0 x Wails 2.0.

.DESCRIPTION
  Abre o app, mede a memória de toda a árvore de processos em repouso e, pela
  interface (UI Automation, os mesmos cliques de um usuário), organiza ~1000
  arquivos com os 6 critérios e desfaz. Pasta do usuário e AppData são
  temporárias: nenhum dado real é tocado.

  A pasta de origem chega à tela pelo registro de "última organização" que o
  app recupera ao abrir. O fim de cada operação é detectado pelo sistema de
  arquivos (origem vazia após organizar; cheia de novo após desfazer).

  Acessibilidade: a UI Automation liga a árvore de acessibilidade, que aumenta a
  memória. No WebView2 ela liga na primeira consulta; o Chromium do Electron só
  a expõe para a UI Automation com --force-renderer-accessibility e
  --enable-features=UiaProvider, ligadas desde a abertura. Por isso
  o repouso "puro" (sem acessibilidade) é medido com -IdleOnly, sem nenhuma
  consulta de UI Automation, e as ações rodam com acessibilidade nas duas versões.

  Electron: a versão 2.0 não traz mais o Electron. Para -App electron, rode este
  script num checkout que ainda o tem (o mesmo das medições em docs/benchmark.md):
    git worktree add ../sortly-electron 22dd905
    cd ../sortly-electron; npm install; npm run build:renderer

.EXAMPLE
  powershell -File scripts/benchmark/memory.ps1 -App wails -IdleOnly
  powershell -File scripts/benchmark/memory.ps1 -App wails
  powershell -File scripts/benchmark/memory.ps1 -App electron -Files 1000
#>
param(
  [Parameter(Mandatory)][ValidateSet('wails', 'electron')][string]$App,
  [int]$Files = 1000,
  [int]$IdleSeconds = 10,
  [switch]$IdleOnly
)

$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName UIAutomationClient, UIAutomationTypes
$root = (Resolve-Path "$PSScriptRoot\..\..").Path
$MB = 1MB

# --- conjunto de arquivos e pasta de usuário temporária -------------------
$work = Join-Path ([IO.Path]::GetTempPath()) ("sortly-bench-$App-" + [guid]::NewGuid().ToString('N').Substring(0, 8))
$home_ = Join-Path $work 'home'
$source = Join-Path $work 'origem'
$destination = Join-Path $work 'destino'
# Estrutura completa de perfil: o Chromium do Electron falha se AppData\Local não existir.
New-Item -ItemType Directory -Force "$home_\AppData\Roaming", "$home_\AppData\Local\Temp", "$home_\.sortly", $source | Out-Null

$fixtures = Get-ChildItem "$root\internal\metadata\testdata\image", "$root\internal\metadata\testdata\mp4", "$root\internal\metadata\testdata\pages" -File
$mtime = Get-Date -Year 2026 -Month 3 -Day 5 -Hour 12 -Minute 0 -Second 0
for ($i = 0; $i -lt $Files; $i++) {
  $f = $fixtures[$i % $fixtures.Count]
  $target = Join-Path $source ('{0}-{1:D4}{2}' -f $f.BaseName, $i, $f.Extension)
  Copy-Item $f.FullName $target
  (Get-Item $target).LastWriteTime = $mtime
}

# Registro com um item fictício: faz o app abrir com origem e destino preenchidos.
$record = @{
  sourceFolderPath      = $source
  destinationFolderPath = $destination
  movedItems            = @(@{ from = "$source\fantasma.txt"; to = "$destination\fantasma.txt" })
  createdFolders        = @()
} | ConvertTo-Json -Compress
[IO.File]::WriteAllText("$home_\.sortly\last-operation.json", $record, (New-Object Text.UTF8Encoding $false))

# --- abrir o app -----------------------------------------------------------
$psi = New-Object Diagnostics.ProcessStartInfo
$psi.UseShellExecute = $false
$psi.EnvironmentVariables['USERPROFILE'] = $home_
$psi.EnvironmentVariables['APPDATA'] = "$home_\AppData\Roaming"
$psi.EnvironmentVariables['LOCALAPPDATA'] = "$home_\AppData\Local"
if ($App -eq 'wails') {
  $psi.FileName = "$root\build\bin\Sortly.exe"
} else {
  $psi.FileName = "$root\node_modules\electron\dist\electron.exe"
  $psi.Arguments = "`"$root`""
  if (-not $IdleOnly) { $psi.Arguments += ' --force-renderer-accessibility --enable-features=UiaProvider' }
  $psi.WorkingDirectory = $root
}
$started = Get-Date
$proc = [Diagnostics.Process]::Start($psi)

# --- amostragem de memória em segundo plano --------------------------------
$samples = [Collections.ArrayList]::Synchronized((New-Object Collections.ArrayList))
$runspace = [runspacefactory]::CreateRunspace(); $runspace.Open()
$runspace.SessionStateProxy.SetVariable('samples', $samples)
$runspace.SessionStateProxy.SetVariable('rootPid', $proc.Id)
$sampler = [powershell]::Create().AddScript({
  while ($true) {
    $all = Get-CimInstance Win32_Process -Property ProcessId, ParentProcessId, PrivatePageCount, WorkingSetSize
    $ids = New-Object 'System.Collections.Generic.HashSet[int]'; [void]$ids.Add($rootPid)
    do { $n = $ids.Count; foreach ($p in $all) { if ($ids.Contains([int]$p.ParentProcessId)) { [void]$ids.Add([int]$p.ProcessId) } } } while ($ids.Count -ne $n)
    $tree = @($all | Where-Object { $ids.Contains([int]$_.ProcessId) })
    [void]$samples.Add([pscustomobject]@{
      T       = Get-Date
      Private = ($tree | Measure-Object PrivatePageCount -Sum).Sum
      WS      = ($tree | Measure-Object WorkingSetSize -Sum).Sum
      Procs   = $tree.Count
    })
    Start-Sleep -Milliseconds 200
  }
})
$sampler.Runspace = $runspace
$null = $sampler.BeginInvoke()

function Summarize($from, $to) {
  $w = @($samples.ToArray() | Where-Object { $_.T -ge $from -and $_.T -le $to })
  if ($w.Count -eq 0) { return $null }
  [ordered]@{
    amostras          = $w.Count
    processos         = ($w | Measure-Object Procs -Maximum).Maximum
    privadaMediaMB    = [math]::Round(($w | Measure-Object Private -Average).Average / $MB, 1)
    privadaPicoMB     = [math]::Round(($w | Measure-Object Private -Maximum).Maximum / $MB, 1)
    workingSetMedioMB = [math]::Round(($w | Measure-Object WS -Average).Average / $MB, 1)
    workingSetPicoMB  = [math]::Round(($w | Measure-Object WS -Maximum).Maximum / $MB, 1)
  }
}

# --- UI Automation ---------------------------------------------------------
function Find-Element([string]$name, $controlType = [Windows.Automation.ControlType]::Button, [int]$timeoutSeconds = 30) {
  $deadline = (Get-Date).AddSeconds($timeoutSeconds)
  $byName = New-Object Windows.Automation.PropertyCondition ([Windows.Automation.AutomationElement]::NameProperty), $name
  $byType = New-Object Windows.Automation.PropertyCondition ([Windows.Automation.AutomationElement]::ControlTypeProperty), $controlType
  $condition = New-Object Windows.Automation.AndCondition $byName, $byType
  while ((Get-Date) -lt $deadline) {
    $proc.Refresh()
    if ($proc.MainWindowHandle -ne 0) {
      $window = [Windows.Automation.AutomationElement]::FromHandle($proc.MainWindowHandle)
      $element = $window.FindFirst([Windows.Automation.TreeScope]::Descendants, $condition)
      if ($element) { return $element }
    }
    Start-Sleep -Milliseconds 250
  }
  throw "Elemento '$name' não encontrado na interface"
}

# O Chromium do Electron às vezes recusa o Invoke ("Erro não reconhecido") quando
# o elemento acabou de ser recriado; tenta de novo com o elemento atualizado.
function Invoke-Element([string]$name) {
  for ($attempt = 1; ; $attempt++) {
    try {
      (Find-Element $name).GetCurrentPattern([Windows.Automation.InvokePattern]::Pattern).Invoke()
      return
    } catch {
      if ($attempt -ge 5) { throw }
      Start-Sleep -Milliseconds 500
    }
  }
}

function Wait-FileCount([int]$expected, [int]$timeoutSeconds = 300) {
  $deadline = (Get-Date).AddSeconds($timeoutSeconds)
  while ((Get-Date) -lt $deadline) {
    if (@(Get-ChildItem $source -File).Count -eq $expected) { return Get-Date }
    Start-Sleep -Milliseconds 100
  }
  throw "Tempo esgotado esperando $expected arquivo(s) na origem"
}

try {
  # Tempo até a janela existir. O repouso é medido antes de qualquer consulta de
  # UI Automation: ela liga a árvore de acessibilidade do Chromium/WebView2, que
  # aumenta a memória. As ações (organizar/desfazer) rodam com ela ligada nas duas versões.
  $deadline = (Get-Date).AddSeconds(60)
  do { Start-Sleep -Milliseconds 50; $proc.Refresh() } while ($proc.MainWindowHandle -eq 0 -and (Get-Date) -lt $deadline)
  $startupMs = [int]((Get-Date) - $started).TotalMilliseconds

  Start-Sleep -Seconds $IdleSeconds
  $idleFrom = Get-Date; Start-Sleep -Seconds 5
  $idle = Summarize $idleFrom (Get-Date)

  if ($IdleOnly) {
    [ordered]@{ app = $App; ateJanelaMs = $startupMs; repousoSemAcessibilidade = $idle } | ConvertTo-Json -Depth 4
    return
  }

  # A tela está pronta quando o botão de organizar aparece (origem já recuperada).
  $null = Find-Element 'Organizar arquivos' -timeoutSeconds 60
  Start-Sleep -Seconds 2
  $a11yFrom = Get-Date; Start-Sleep -Seconds 3
  $a11y = Summarize $a11yFrom (Get-Date)

  # Liga os 6 critérios no painel de configurações.
  Invoke-Element 'Configurações de organização'
  foreach ($label in 'Duração (.mp4)', 'Páginas', 'Resolução', 'Data', 'Tamanho (MB)') {
    $box = Find-Element $label ([Windows.Automation.ControlType]::CheckBox)
    $toggle = $box.GetCurrentPattern([Windows.Automation.TogglePattern]::Pattern)
    if ($toggle.Current.ToggleState -ne [Windows.Automation.ToggleState]::On) { $toggle.Toggle() }
  }
  Invoke-Element 'Configurações de organização'
  Start-Sleep -Milliseconds 500

  $orgFrom = Get-Date
  Invoke-Element 'Organizar arquivos'
  $orgTo = Wait-FileCount 0
  Start-Sleep -Seconds 2

  $undoFrom = Get-Date
  Invoke-Element 'Desfazer ultima separação'
  $undoTo = Wait-FileCount $Files
  Start-Sleep -Seconds 2

  [ordered]@{
    app              = $App
    arquivos         = $Files
    ateJanelaMs      = $startupMs
    repousoComAcessibilidade = $a11y
    organizar        = [ordered]@{ duracaoMs = [int]($orgTo - $orgFrom).TotalMilliseconds } + (Summarize $orgFrom $orgTo.AddMilliseconds(200))
    desfazer         = [ordered]@{ duracaoMs = [int]($undoTo - $undoFrom).TotalMilliseconds } + (Summarize $undoFrom $undoTo.AddMilliseconds(200))
  } | ConvertTo-Json -Depth 4
}
finally {
  $sampler.Stop(); $runspace.Close()
  cmd /c "taskkill /PID $($proc.Id) /T /F >nul 2>&1"
  Start-Sleep -Seconds 1
  # O WebView2 pode segurar arquivos por alguns segundos após fechar.
  for ($try = 0; $try -lt 10 -and (Test-Path -LiteralPath $work); $try++) {
    Remove-Item -LiteralPath $work -Recurse -Force -ErrorAction SilentlyContinue
    if (Test-Path -LiteralPath $work) { Start-Sleep -Seconds 1 }
  }
}
