**Запуск**

```
go run ./cmd/server
```

**Сборка**

*Для Linux*

```powershell
$Env:GOOS = "linux"; $Env:GOARCH = "amd64"

go build -o ../distr ./cmd/server
```