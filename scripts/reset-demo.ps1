# Zera os bancos e recria o ambiente do zero, para gravar a demonstração
# com dados limpos.
#
# Uso:  .\scripts\reset-demo.ps1

$root = Split-Path $PSScriptRoot -Parent
Push-Location $root

Write-Host "Derrubando containers e apagando os volumes dos bancos..." -ForegroundColor Cyan
# O docker compose escreve o progresso em stderr; o 2>&1 evita que o
# PowerShell interprete essas linhas como erro.
docker compose down -v 2>&1 | Out-Null

Write-Host "Subindo tudo novamente..." -ForegroundColor Cyan
# O --build garante que o que sobe reflete o código atual: sem ele, uma
# alteração no frontend não apareceria e a demonstração rodaria em cima
# de uma imagem antiga.
docker compose up -d --build 2>&1 | Out-Null

Pop-Location

Write-Host "Aguardando os servicos ficarem prontos..." -ForegroundColor Cyan
$pronto = $false
for ($i = 0; $i -lt 30; $i++) {
    try {
        Invoke-RestMethod -Uri "http://localhost:8081/health" -TimeoutSec 2 | Out-Null
        Invoke-RestMethod -Uri "http://localhost:8082/health" -TimeoutSec 2 | Out-Null
        $pronto = $true
        break
    } catch {
        Start-Sleep -Seconds 2
    }
}

Write-Host ""
if ($pronto) {
    Write-Host "Ambiente limpo e pronto para gravar!" -ForegroundColor Green
    Write-Host "  Frontend:    http://localhost:4200"
    Write-Host "  Estoque:     http://localhost:8081"
    Write-Host "  Faturamento: http://localhost:8082"
} else {
    Write-Host "Os servicos nao responderam a tempo. Rode 'docker compose ps' para verificar." -ForegroundColor Red
}
