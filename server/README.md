**Запуск**

```
go run ./cmd/server

или для отслеживания изменений файлов

air
```

**Сборка**

*Для Linux*

```powershell
$Env:GOOS = "linux"; $Env:GOARCH = "amd64"

go build -o ../dist ./cmd/server
```