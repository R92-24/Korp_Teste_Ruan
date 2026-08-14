# Demonstra o tratamento de concorrência: um produto com saldo 1 sendo
# disputado por duas notas fiscais que são impressas ao mesmo tempo.
#
# Resultado esperado: uma nota fecha, a outra recebe 409 SALDO_INSUFICIENTE,
# e o saldo do produto termina em 0 (nunca negativo).
#
# Uso:  .\scripts\demo-concorrencia.ps1

$ErrorActionPreference = "Stop"
$estoque = "http://localhost:8081"
$faturamento = "http://localhost:8082"
$codigo = "CONC-" + (Get-Random -Maximum 9999)

Write-Host "1) Criando produto '$codigo' com saldo 1..." -ForegroundColor Cyan
$produto = Invoke-RestMethod -Uri "$estoque/produtos" -Method Post -ContentType "application/json" `
    -Body (@{ codigo = $codigo; descricao = "Produto Disputado"; saldo = 1 } | ConvertTo-Json)
Write-Host "   Produto criado com saldo = $($produto.saldo)" -ForegroundColor Green

Write-Host "2) Criando duas notas, cada uma pedindo 1 unidade desse produto..." -ForegroundColor Cyan
$notaA = Invoke-RestMethod -Uri "$faturamento/notas" -Method Post -ContentType "application/json" -Body '{}'
$notaB = Invoke-RestMethod -Uri "$faturamento/notas" -Method Post -ContentType "application/json" -Body '{}'
$itemBody = (@{ codigo = $codigo; quantidade = 1 } | ConvertTo-Json)
Invoke-RestMethod -Uri "$faturamento/notas/$($notaA.numero)/itens" -Method Post -ContentType "application/json" -Body $itemBody | Out-Null
Invoke-RestMethod -Uri "$faturamento/notas/$($notaB.numero)/itens" -Method Post -ContentType "application/json" -Body $itemBody | Out-Null
Write-Host "   Nota A = $($notaA.numero) | Nota B = $($notaB.numero)" -ForegroundColor Green

Write-Host "3) Imprimindo AS DUAS ao mesmo tempo..." -ForegroundColor Cyan
$imprimir = {
    param($url)
    try {
        $r = Invoke-RestMethod -Uri $url -Method Post -ContentType "application/json" -Body '{}'
        "SUCESSO - nota $($r.numero) ficou $($r.status)"
    } catch {
        "REJEITADA - $($_.Exception.Response.StatusCode) (saldo insuficiente)"
    }
}
$jobA = Start-Job -ScriptBlock $imprimir -ArgumentList "$faturamento/notas/$($notaA.numero)/imprimir"
$jobB = Start-Job -ScriptBlock $imprimir -ArgumentList "$faturamento/notas/$($notaB.numero)/imprimir"
Wait-Job $jobA, $jobB | Out-Null

Write-Host ""
Write-Host "   Nota A ($($notaA.numero)): $(Receive-Job $jobA)" -ForegroundColor Yellow
Write-Host "   Nota B ($($notaB.numero)): $(Receive-Job $jobB)" -ForegroundColor Yellow
Remove-Job $jobA, $jobB

$final = Invoke-RestMethod -Uri "$estoque/produtos/$codigo" -Method Get
Write-Host ""
Write-Host "4) Saldo final do produto: $($final.saldo)  (esperado: 0, nunca negativo)" -ForegroundColor Green
