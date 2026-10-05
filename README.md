[Запуск клиента](./client/README.md)

## Установка сервера

Скачайте и запустите скрипт установки на сервере (Ubuntu/Debian):

```bash
curl -fsSL https://raw.githubusercontent.com/MetallIdea/adpanel/refs/heads/main/dist/install.sh | sudo bash
```

Скрипт выполнит следующие действия:

- обновит систему и установит зависимости (curl, wget, nginx, unzip, ufw);
- скачает и установит фронтенд в `/var/www/adpanel`;
- скачает и установит бинарный файл сервера в `/opt/adpanel/adpanel-server`;
- скачает скрипты управления (`start.sh`, `stop.sh`, `update.sh`) в `/opt/adpanel`;
- создаст systemd-сервис `adpanel-server`;
- настроит nginx для проксирования запросов;
- настроит фаервол (ufw);
- сгенерирует учетные данные администратора (логин: `admin`, пароль — случайная строка из 16 символов).

После установки:

```bash
# Проверка статуса сервера
systemctl status adpanel-server

# Просмотр логов
journalctl -u adpanel-server -f

# Управление
/opt/adpanel/start.sh
/opt/adpanel/stop.sh
/opt/adpanel/update.sh
```

Доступ к приложению: `http://<IP_сервера>:8765`