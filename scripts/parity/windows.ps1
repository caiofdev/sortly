<#
.SYNOPSIS
  Roteiro automatizado de paridade do Sortly no Windows (itens manuais na issue #40).

.DESCRIPTION
  Usa a interface do app (UI Automation) com uma pasta de usuário temporária e
  confere o resultado no sistema de arquivos. Cobre os itens que dá para
  automatizar; arrastar do Explorer e o seletor de pastas continuam manuais.

  -OtherVolume: pasta em outro disco para o teste de mover entre volumes (B3).
  Uma subpasta temporária é criada lá e removida no fim.

.EXAMPLE
  powershell -File scripts/parity/windows.ps1 -OtherVolume E:\
#>
param(
  [string]$Exe = "$PSScriptRoot\..\..\build\bin\Sortly.exe",
  [string]$OtherVolume = ''
)

$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName UIAutomationClient, UIAutomationTypes
$root = (Resolve-Path "$PSScriptRoot\..\..").Path
$Exe = (Resolve-Path $Exe).Path
$results = [ordered]@{}
$UIA = [Windows.Automation.AutomationElement]

$work = Join-Path ([IO.Path]::GetTempPath()) ('sortly-parity-' + [guid]::NewGuid().ToString('N').Substring(0, 8))
$home_ = Join-Path $work 'home'
New-Item -ItemType Directory -Force "$home_\AppData\Roaming", "$home_\AppData\Local\Temp", "$home_\.sortly" | Out-Null
$recordPath = "$home_\.sortly\last-operation.json"
$script:proc = $null

function Check([string]$id, [bool]$ok, [string]$detail = '') {
  $results[$id] = [ordered]@{ ok = $ok; detalhe = $detail }
  Write-Host ('{0} {1} {2}' -f ($(if ($ok) { 'OK  ' } else { 'FALHA' })), $id, $detail)
}

function New-Dataset([string]$dir) {
  New-Item -ItemType Directory -Force $dir, "$dir\subpasta" | Out-Null
  $mtime = Get-Date -Year 2026 -Month 3 -Day 5 -Hour 12 -Minute 0 -Second 0
  foreach ($group in 'image', 'mp4', 'pages') {
    Get-ChildItem "$root\backend\metadata\testdata\$group" -File | ForEach-Object { Copy-Item $_.FullName $dir }
  }
  Set-Content "$dir\LEIAME" 'sem extensão'; Set-Content "$dir\.gitignore" 'oculto'; Set-Content "$dir\nota.txt" 'texto'
  Set-Content "$dir\subpasta\dentro.txt" 'não deve ser tocado'
  Get-ChildItem $dir -File -Force | ForEach-Object { $_.LastWriteTime = $mtime }
}

# Registro com um item fictício: o app abre com origem e destino preenchidos.
function Set-Record([string]$source, [string]$destination) {
  $json = @{ sourceFolderPath = $source; destinationFolderPath = $destination
    movedItems = @(@{ from = "$source\fantasma.txt"; to = "$destination\fantasma.txt" }); createdFolders = @() } | ConvertTo-Json -Compress
  [IO.File]::WriteAllText($recordPath, $json, (New-Object Text.UTF8Encoding $false))
}

function Start-App {
  $psi = New-Object Diagnostics.ProcessStartInfo $Exe
  $psi.UseShellExecute = $false
  $psi.EnvironmentVariables['USERPROFILE'] = $home_
  $psi.EnvironmentVariables['APPDATA'] = "$home_\AppData\Roaming"
  $psi.EnvironmentVariables['LOCALAPPDATA'] = "$home_\AppData\Local"
  $script:proc = [Diagnostics.Process]::Start($psi)
  $null = Find-Element 'PT-BR', 'EN' 60
}

# Fecha a janela como o usuário (o WebView2 grava o localStorage ao fechar);
# só força o encerramento se o app não fechar sozinho.
function Stop-App {
  if (-not $script:proc) { return }
  [void]$script:proc.CloseMainWindow()
  if (-not $script:proc.WaitForExit(10000)) { cmd /c "taskkill /PID $($script:proc.Id) /T /F >nul 2>&1" }
  Start-Sleep -Seconds 2
  $script:proc = $null
}

# -Last pega a última ocorrência do nome: "Organizar" é o item da sidebar e
# também a ação da página, que vem depois na árvore (#74).
function Find-Element([string[]]$names, [int]$timeoutSeconds = 20, $type = $null, [switch]$Last) {
  $deadline = (Get-Date).AddSeconds($timeoutSeconds)
  while ((Get-Date) -lt $deadline) {
    $script:proc.Refresh()
    if ($script:proc.MainWindowHandle -ne 0) {
      $window = $UIA::FromHandle($script:proc.MainWindowHandle)
      foreach ($name in $names) {
        $condition = New-Object Windows.Automation.PropertyCondition ($UIA::NameProperty), $name
        if ($type) { $condition = New-Object Windows.Automation.AndCondition $condition, (New-Object Windows.Automation.PropertyCondition ($UIA::ControlTypeProperty), $type) }
        $found = @($window.FindAll([Windows.Automation.TreeScope]::Descendants, $condition))
        if ($found.Count -gt 0) { if ($Last) { return $found[-1] } else { return $found[0] } }
      }
    }
    Start-Sleep -Milliseconds 250
  }
  throw "Elemento '$($names -join "' / '")' não encontrado"
}

# Trocar de página recria os elementos: um elemento achado no meio da troca
# deixa de expor os padrões, então a busca é refeita até o clique dar certo (#74).
function Click([string[]]$name, [switch]$Last) {
  for ($try = 0; $try -lt 10; $try++) {
    $element = Find-Element $name -type ([Windows.Automation.ControlType]::Button) -Last:$Last
    $pattern = $null
    # Botões com aria-pressed (idioma) são de alternância: expõem Toggle, não Invoke (#60).
    if ($element.TryGetCurrentPattern([Windows.Automation.InvokePattern]::Pattern, [ref]$pattern)) { $pattern.Invoke(); return }
    if ($element.TryGetCurrentPattern([Windows.Automation.TogglePattern]::Pattern, [ref]$pattern)) { $pattern.Toggle(); return }
    # O sino tem aria-expanded: expõe ExpandCollapse, também sem Invoke (#74).
    if ($element.TryGetCurrentPattern([Windows.Automation.ExpandCollapsePattern]::Pattern, [ref]$pattern)) {
      if ($pattern.Current.ExpandCollapseState -eq 'Expanded') { $pattern.Collapse() } else { $pattern.Expand() }
      return
    }
    Start-Sleep -Milliseconds 300
  }
  throw "Botão '$($name -join "' / '")' não aceita clique"
}

function Toggle-State([string[]]$name) {
  (Find-Element $name -type ([Windows.Automation.ControlType]::Button)).GetCurrentPattern([Windows.Automation.TogglePattern]::Pattern).Current.ToggleState
}

function Is-Enabled([string[]]$name, [switch]$Last) { (Find-Element $name -type ([Windows.Automation.ControlType]::Button) -Last:$Last).Current.IsEnabled }

# Os critérios são switches (role="switch") na página Configurações (#74).
function Criterion([string]$label) {
  (Find-Element $label -type ([Windows.Automation.ControlType]::Button)).GetCurrentPattern([Windows.Automation.TogglePattern]::Pattern)
}

function Set-Criteria([string[]]$on) {
  Click 'Configurações'
  foreach ($label in 'Tipo', 'Duração (.mp4)', 'Páginas', 'Resolução', 'Data', 'Tamanho (MB)', 'Extensão do arquivo') {
    $toggle = Criterion $label
    $want = $on -contains $label
    if (($toggle.Current.ToggleState -eq 'On') -ne $want -and $want) { $toggle.Toggle(); Start-Sleep -Milliseconds 200 }
  }
  foreach ($label in 'Tipo', 'Duração (.mp4)', 'Páginas', 'Resolução', 'Data', 'Tamanho (MB)', 'Extensão do arquivo') {
    $toggle = Criterion $label
    if ($toggle.Current.ToggleState -eq 'On' -and $on -notcontains $label) { $toggle.Toggle(); Start-Sleep -Milliseconds 200 }
  }
  Click 'Organizar'
}

# Depois de organizar, a tela é o Concluído, sem o botão Organizar; reabrir volta
# à tela inicial com a origem do último registro (#79).
function Restart-App {
  Stop-App
  Start-App
}

# Primeiro texto da janela que casa com o padrão; o painel vem antes dos toasts (#75).
function Find-Text([string]$pattern) {
  $window = $UIA::FromHandle($script:proc.MainWindowHandle)
  $texts = $window.FindAll([Windows.Automation.TreeScope]::Descendants, (New-Object Windows.Automation.PropertyCondition ($UIA::ControlTypeProperty), ([Windows.Automation.ControlType]::Text)))
  return [string]($texts | ForEach-Object { $_.Current.Name } | Where-Object { $_ -match $pattern } | Select-Object -First 1)
}

function Last-Notice {
  Click 'Notificações', 'Notifications'
  Start-Sleep -Milliseconds 600
  # Só textos de aviso, do mais recente para o mais antigo (#75).
  $notice = Find-Text 'organizad|organized|Interrompido|Stopped|com falha|failed|não foi encontrado|não foram encontrados|not found|Invalid folder|Pasta inválida|Nenhuma separação|No recent'
  Click 'Notificações', 'Notifications'
  return [string]$notice
}

function Wait-Until([scriptblock]$condition, [int]$timeoutSeconds = 60) {
  $deadline = (Get-Date).AddSeconds($timeoutSeconds)
  while ((Get-Date) -lt $deadline) { if (& $condition) { return $true }; Start-Sleep -Milliseconds 200 }
  return $false
}

function Tree([string]$dir) { @(Get-ChildItem $dir -Recurse -File -Force | ForEach-Object { $_.FullName.Substring($dir.Length + 1) } | Sort-Object) }

$otherDir = $null
try {
  # --- 1) Todos os critérios, destino em outro volume (B3), desfazer após reabrir ---
  $src = "$work\origem"; New-Dataset $src
  $before = Tree $src
  $dst = if ($OtherVolume) { $otherDir = Join-Path $OtherVolume ('sortly-parity-' + [guid]::NewGuid().ToString('N').Substring(0, 8)); $otherDir } else { "$work\destino" }
  Set-Record $src $dst
  Start-App
  Set-Criteria @('Duração (.mp4)', 'Páginas', 'Resolução', 'Data', 'Tamanho (MB)', 'Extensão do arquivo')
  $found = Wait-Until { (Find-Text '^arquivos? encontrados?$') -ne '' } 15
  Check 'WIN_PREVIEW' ($found -and (Find-Text '^pdf$') -ne '') 'prévia da origem recuperada, com o chip da pasta pdf'
  Click 'Organizar' -Last
  $moved = Wait-Until { (Tree $src).Count -eq 3 }
  $organized = Tree $dst
  $toast = Find-Text '^Pronto! \d+ arquivos organizados$'
  Check 'WIN_TOAST' ($toast -ne '') "toast ao organizar: $toast"
  Check 'WIN_CRIT_ALL' ($moved -and $organized -contains 'pdf\date-2026-03-05\size-1mb\pages-3\3-pages-classic-xref.pdf') 'pdf/date-*/size-*/pages-3'
  Check 'WIN_CRIT_EXT' ($organized -contains 'txt\date-2026-03-05\size-1mb\nota.txt') 'txt/…/nota.txt; LEIAME e .gitignore ficam na origem'
  Check 'WIN_CRIT_RES' ($organized -contains 'png\date-2026-03-05\size-1mb\3x2\3x2.png') 'png/…/3x2'
  Check 'WIN_CRIT_DUR' ($organized -contains 'mp4\date-2026-03-05\size-1mb\duration-00h00m59s\video-only-59s.mp4') 'vídeo sem áudio → duration-00h00m59s'
  Check 'WIN_CRIT_PAGES' (($organized -contains 'docx\date-2026-03-05\size-1mb\pages-12\12-pages.docx') -and ($organized -contains 'doc\date-2026-03-05\size-1mb\pages-unknown\legacy.doc')) 'docx pages-12; .doc pages-unknown'
  if ($OtherVolume) { Check 'WIN_B3' ($moved -and (Split-Path -Qualifier $dst) -ne (Split-Path -Qualifier $src)) "$((Split-Path -Qualifier $src)) → $((Split-Path -Qualifier $dst))" }
  Stop-App

  Start-App
  $canUndo = Is-Enabled 'Desfazer'
  Click 'Desfazer'
  $restored = Wait-Until { (Compare-Object $before (Tree $src)) -eq $null }
  Check 'WIN_UNDO_RESTART' ($canUndo -and $restored) 'reaberto: desfazer ativo e árvore original restaurada'
  Check 'WIN_UNDO' ($restored -and -not (Test-Path $dst) -or @(Get-ChildItem $dst -Recurse -File -ErrorAction SilentlyContinue).Count -eq 0) 'pastas criadas removidas do destino'
  # O histórico sobrevive ao reabrir, e o desfazer marca a organização como desfeita (#80).
  Click 'Histórico'
  $historyRow = (Find-Text "^$([regex]::Escape($src))$") -ne '' -and (Find-Text '^Desfeito$') -ne ''
  Check 'WIN_HISTORY' $historyRow 'histórico com a origem e o status Desfeito depois de reabrir e desfazer'
  Click 'Organizar'

  # --- 2) Último critério não pode ser desmarcado ---
  Set-Criteria @('Extensão do arquivo')
  Click 'Configurações'
  $last = Find-Element 'Extensão do arquivo' -type ([Windows.Automation.ControlType]::Button)
  Check 'WIN_CRIT_LAST' (-not $last.Current.IsEnabled) 'switch do último critério desabilitado'
  Click 'Organizar'

  # --- 3) Conflito de nome (B4) e arquivo bloqueado (B1) ---
  New-Item -ItemType Directory -Force "$dst\txt" | Out-Null; Set-Content "$dst\txt\nota.txt" 'já existia'
  $lock = [IO.File]::Open("$src\3x2.png", 'Open', 'Read', 'None')
  try {
    Click 'Organizar' -Last
    $done = Wait-Until { (Tree $src).Count -eq 4 }   # LEIAME, .gitignore, subpasta\dentro.txt e o png bloqueado
    $notice = Last-Notice
  } finally { $lock.Close() }
  Check 'WIN_CONFLICT' ((Get-Content "$dst\txt\nota.txt") -eq 'já existia' -and (Test-Path "$dst\txt\nota (1).txt")) 'nota.txt existente preservado; novo vira nota (1).txt'
  Check 'WIN_B1' ($done -and $notice -match ', 1 com falha$') $notice
  $doneTitle = Find-Text '^Pronto! Arquivos? organizados?\.$'
  $doneOpen = $null -ne (Find-Element 'Abrir pasta de destino' 5)
  Check 'WIN_DONE' ($doneTitle -ne '' -and $doneOpen) "Concluído: $doneTitle"
  Click 'Organizar outra pasta'
  $startOver = (Find-Text '^Nenhuma pasta selecionada$') -ne '' -and (Is-Enabled 'Desfazer')
  Check 'WIN_START_OVER' $startOver 'Organizar outra pasta volta ao início sem origem, com o desfazer ativo'

  # --- 4) Organizar sem nada para mover mantém o desfazer (B2) ---
  Restart-App
  Click 'Organizar' -Last; Start-Sleep -Seconds 2
  Restart-App
  Click 'Organizar' -Last; Start-Sleep -Seconds 2   # nada para mover
  $keep = Is-Enabled 'Desfazer'
  Check 'WIN_B2' $keep 'botão de desfazer continua ativo após organizar sem movimentos'

  # --- 5) Arquivo apagado antes de desfazer ---
  Remove-Item "$dst\png\3x2.png" -ErrorAction SilentlyContinue
  Click 'Desfazer'; Start-Sleep -Seconds 2
  $notice = Last-Notice
  Check 'WIN_UNDO_MISSING' ($notice -match '1 não foi encontrado\.') $notice

  # --- 6) B8: só "Resolução" na própria pasta não renomeia os demais ---
  Stop-App
  $b8 = "$work\b8"; New-Dataset $b8; $b8Before = Tree $b8
  Set-Record $b8 $b8
  Start-App
  Set-Criteria @('Resolução')
  Click 'Organizar' -Last; Start-Sleep -Seconds 3
  $after = Tree $b8
  $renamed = @($after | Where-Object { $_ -match ' \(1\)' })
  Check 'WIN_B8' ($renamed.Count -eq 0 -and ($after -contains '3x2\3x2.png')) "renomeados: $($renamed.Count)"

  # --- 6b) Cancelar no meio: o que foi movido fica no registro e pode ser desfeito (#78) ---
  Stop-App
  $big = "$work\cancelar"; New-Item -ItemType Directory -Force $big | Out-Null
  foreach ($i in 1..4000) { [IO.File]::WriteAllText("$big\arquivo-$i.txt", 'x') }
  Set-Record $big $big
  Start-App
  Set-Criteria @('Extensão do arquivo')
  Click 'Organizar' -Last
  # Cancela só depois do primeiro movimento, para conferir o registro parcial.
  $moving = Wait-Until { @(Get-ChildItem $big -File).Count -lt 4000 } 30
  Click 'Cancelar'
  Start-Sleep -Seconds 2
  $left = @(Get-ChildItem $big -File).Count
  $notice = Last-Notice
  Click 'Desfazer'
  $back = Wait-Until { @(Get-ChildItem $big -File).Count -eq 4000 }
  Check 'WIN_CANCEL' ($moving -and $notice -match '^Interrompido: \d+ arquivos movidos$' -and $left -gt 0 -and $left -lt 4000 -and $back) "$notice; movendo: $moving; ficaram $left na origem; desfeito: $back"

  # --- 6b2) Desfazer com progresso e Cancelar: o que não voltou continua desfazível (#96) ---
  Click 'Organizar' -Last
  $organized = Wait-Until { @(Get-ChildItem $big -File).Count -eq 0 } 60
  Click 'Desfazer'
  $returning = Wait-Until { @(Get-ChildItem $big -File).Count -gt 0 } 30
  $screen = Find-Text '^Desfazendo…$'
  Click 'Cancelar'
  Start-Sleep -Seconds 2
  $returned = @(Get-ChildItem $big -File).Count
  $notice = Last-Notice
  Click 'Desfazer' -Last
  $back = Wait-Until { @(Get-ChildItem $big -File).Count -eq 4000 }
  Check 'WIN_UNDO_CANCEL' ($organized -and $returning -and $screen -ne '' -and $notice -match '^Desfazer interrompido: \d+ arquivos voltaram$' -and $returned -gt 0 -and $returned -lt 4000 -and $back) "$notice; tela de progresso: '$screen'; voltaram $returned; desfeito depois: $back"

  # --- 6c) Critério Tipo: categorias no 1º nível e o sem extensão em other (#81) ---
  Stop-App
  $typ = "$work\tipo"; New-Dataset $typ
  Set-Record $typ $typ
  Start-App
  Set-Criteria @('Tipo')
  Click 'Organizar' -Last
  $typed = Wait-Until { Test-Path "$typ\other\LEIAME" }
  $tree = Tree $typ
  Check 'WIN_CRIT_TYPE' ($typed -and $tree -contains 'images\3x2.png' -and $tree -contains 'documents\nota.txt' -and $tree -contains 'videos\video-only-59s.mp4') 'images, documents, videos e LEIAME em other'

  # --- 6d) Duplicados: Ignorar deixa na origem; Substituir guarda o antigo e o desfazer devolve os dois (#82) ---
  Stop-App
  $dup = "$work\dup"; $dupDst = "$work\dup-destino"
  New-Item -ItemType Directory -Force $dup, "$dupDst\txt" | Out-Null
  Set-Content "$dup\nota.txt" 'nova'; Set-Content "$dupDst\txt\nota.txt" 'antiga'
  Set-Record $dup $dupDst
  Start-App
  Set-Criteria @('Extensão do arquivo')
  Click 'Configurações'; Click 'Ignorar'; Start-Sleep -Milliseconds 500; Click 'Organizar'
  Click 'Organizar' -Last; Start-Sleep -Seconds 2
  $skipped = ((Get-Content "$dup\nota.txt") -eq 'nova') -and ((Get-Content "$dupDst\txt\nota.txt") -eq 'antiga')
  Check 'WIN_DUP_SKIP' $skipped 'Ignorar: nota.txt fica na origem e o existente continua no destino'
  Click 'Configurações'; Click 'Substituir'; Start-Sleep -Milliseconds 500; Click 'Organizar'
  Click 'Organizar' -Last; Start-Sleep -Seconds 2
  $backupFile = @(Get-ChildItem "$home_\.sortly\substituidos" -Recurse -File -ErrorAction SilentlyContinue | Select-Object -First 1)
  $replaced = ((Get-Content "$dupDst\txt\nota.txt") -eq 'nova') -and $backupFile.Count -eq 1 -and ((Get-Content $backupFile[0].FullName) -eq 'antiga')
  Click 'Desfazer'
  # No meio do desfazer, o destino fica sem nota.txt até o substituído voltar do backup (#84).
  $restored = Wait-Until { (Test-Path "$dup\nota.txt") -and ((Get-Content "$dupDst\txt\nota.txt" -ErrorAction SilentlyContinue) -eq 'antiga') } 15
  $cleaned = @(Get-ChildItem "$home_\.sortly\substituidos" -Recurse -File -ErrorAction SilentlyContinue).Count -eq 0
  Check 'WIN_DUP_REPLACE' ($replaced -and $restored -and $cleaned) "substituído e guardado: $replaced; desfeito: $restored; backup apagado: $cleaned"
  Click 'Configurações'; Click 'Renomear'; Start-Sleep -Milliseconds 500; Click 'Organizar'

  # --- 6e) Incluir subpastas: subpasta\dentro.txt vai para txt e a subpasta vazia some; o desfazer a recria (#83) ---
  Stop-App
  $sub = "$work\subpastas"; New-Dataset $sub
  Set-Record $sub $sub
  Start-App
  Set-Criteria @('Extensão do arquivo')
  Click 'Configurações'; Click 'Incluir subpastas'; Start-Sleep -Milliseconds 500; Click 'Organizar'
  Click 'Organizar' -Last
  $flattened = Wait-Until { (Test-Path "$sub\txt\dentro.txt") -and -not (Test-Path "$sub\subpasta") } 15
  Click 'Desfazer'
  $recreated = Wait-Until { Test-Path "$sub\subpasta\dentro.txt" } 15
  Check 'WIN_SUBFOLDERS' ($flattened -and $recreated) "achatado: $flattened; subpasta recriada no desfazer: $recreated"
  Click 'Configurações'; Click 'Incluir subpastas'; Start-Sleep -Milliseconds 500; Click 'Organizar'

  # O desfazer apagou o registro; o passo 7 reabre o app e precisa da origem do B8.
  Set-Criteria @('Resolução')
  Stop-App
  Set-Record $b8 $b8
  Start-App

  # --- 7) Idioma, preferências e erro em inglês (B6) ---
  $on = [Windows.Automation.ToggleState]::On
  $off = [Windows.Automation.ToggleState]::Off
  Check 'WIN_A11Y_LANG' ((Toggle-State 'PT-BR') -eq $on -and (Toggle-State 'EN') -eq $off) 'leitor de tela: botão "PT-BR" marcado, "EN" não (aria-pressed)'
  Click 'Configurações'; Click 'Claro'; Start-Sleep -Milliseconds 500; Click 'Organizar'
  Click 'EN'; Start-Sleep -Milliseconds 500
  Stop-App
  Start-App
  $en = $null -ne (Find-Element 'Organize folder' 10)
  Click 'Settings'
  $resOn = (Criterion 'Resolution').Current.ToggleState -eq 'On'
  $light = (Toggle-State 'Light') -eq $on -and (Toggle-State 'Dark') -eq $off
  Click 'Organize'
  Check 'WIN_LANG' $en 'botões em inglês após trocar o idioma'
  Check 'WIN_PREFS' ($en -and $resOn) 'idioma e critério "Resolution" mantidos após reabrir'
  Check 'WIN_THEME' $light 'tema claro mantido após reabrir'
  # As preferências ficam em ~/.sortly/settings.json, fora do WebView (#44):
  # apagar AppData (dados do WebView2 e localStorage) não pode perdê-las.
  Stop-App
  for ($try = 0; $try -lt 10 -and (Test-Path -LiteralPath "$home_\AppData"); $try++) {
    Remove-Item -LiteralPath "$home_\AppData" -Recurse -Force -ErrorAction SilentlyContinue
    if (Test-Path -LiteralPath "$home_\AppData") { Start-Sleep -Seconds 1 }
  }
  New-Item -ItemType Directory -Force "$home_\AppData\Roaming", "$home_\AppData\Local\Temp" | Out-Null
  Start-App
  $enAgain = $null -ne (Find-Element 'Organize folder' 10)
  Click 'Settings'
  $resAgain = (Criterion 'Resolution').Current.ToggleState -eq 'On'
  Click 'Organize'
  Check 'WIN_PREFS_WEBVIEW' ($enAgain -and $resAgain) 'idioma e critério mantidos após apagar os dados do WebView'
  # Limpar as notificações devolve uma lista vazia; a tela precisa continuar lá (#45).
  Click 'Notifications'; Start-Sleep -Milliseconds 500
  Click 'Clear'; Start-Sleep -Milliseconds 500
  Click 'Notifications'; Start-Sleep -Milliseconds 500
  Check 'WIN_CLEAR' ($null -ne (Find-Element 'Organize folder' 5)) 'tela continua após limpar as notificações'
  Rename-Item $b8 "$work\b8-renomeada"
  Click 'Organize' -Last; Start-Sleep -Seconds 2
  $notice = Last-Notice
  Check 'WIN_B6' ($notice -eq 'Invalid folder.') $notice
}
finally {
  Stop-App
  if ($otherDir) { Remove-Item -LiteralPath $otherDir -Recurse -Force -ErrorAction SilentlyContinue }
  # O WebView2 pode segurar arquivos por alguns segundos após fechar.
  for ($try = 0; $try -lt 10 -and (Test-Path -LiteralPath $work); $try++) {
    Remove-Item -LiteralPath $work -Recurse -Force -ErrorAction SilentlyContinue
    if (Test-Path -LiteralPath $work) { Start-Sleep -Seconds 1 }
  }
  $results | ConvertTo-Json -Depth 3
}
