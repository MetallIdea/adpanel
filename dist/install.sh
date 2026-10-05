#!/bin/bash
set -e

# ============================
# Конфигурация
# ============================
APP_DIR="/opt/adpanel"
FRONTEND_URL="https://github.com/MetallIdea/adpanel/raw/refs/heads/main/dist/frontend.zip"
START_URL="https://raw.githubusercontent.com/MetallIdea/adpanel/refs/heads/main/dist/start.sh"
STOP_URL="https://raw.githubusercontent.com/MetallIdea/adpanel/refs/heads/main/dist/stop.sh"
UPDATE_URL="https://raw.githubusercontent.com/MetallIdea/adpanel/refs/heads/main/dist/update.sh"
BINARY_URL="https://raw.githubusercontent.com/MetallIdea/adpanel/refs/heads/main/dist/server"  # URL для скачивания дистрибутива сервера
DOMAIN=""      # Укажите домен или оставьте пустым
PORT=8080      # Порт приложения
EXTERNAL_PORT=8765

# ============================
# 1. Установка зависимостей
# ============================
echo "=== Установка зависимостей ==="
apt install -y \
    wget \
    nginx \
    unzip \
    ufw

# ============================
# 3. Создание директории приложения
# ============================
echo "=== Подготовка директории ==="
mkdir -p "$APP_DIR"
mkdir -p /var/www/adpanel

# ============================
# 4. Скачивание и установка фронтенда
# ============================
echo "=== Установка фронтенда ==="

if [ -n "$FRONTEND_URL" ]; then
    TMP_FRONTEND=$(mktemp /tmp/adpanel-frontend-XXXXXX.zip)
    echo "Скачивание фронтенда: $FRONTEND_URL"
    wget -qO "$TMP_FRONTEND" "$FRONTEND_URL"
    unzip -qo "$TMP_FRONTEND" -d /var/www/adpanel
    rm -f "$TMP_FRONTEND"
    echo "Фронтенд установлен: /var/www/adpanel"
else
    echo "ПРЕДУПРЕЖДЕНИЕ: FRONTEND_URL не задан, фронтенд не установлен"
fi

# ============================
# 5. Скачивание и установка сервера
# ============================
echo "=== Установка сервера ==="

if [ -z "$BINARY_URL" ]; then
    echo "ОШИБКА: BINARY_URL не задан"
    exit 1
fi

BINARY_PATH="$APP_DIR/adpanel-server"
TMP_BINARY=$(mktemp /tmp/adpanel-server-XXXXXX)

echo "Скачивание: $BINARY_URL"
wget -qO "$TMP_BINARY" "$BINARY_URL"
chmod +x "$TMP_BINARY"
mv "$TMP_BINARY" "$BINARY_PATH"

echo "Сервер установлен: $BINARY_PATH"

# ============================
# 5. Скачивание скриптов управления
# ============================
echo "=== Установка скриптов управления ==="

TMP_START=$(mktemp /tmp/adpanel-start-XXXXXX.sh)
TMP_STOP=$(mktemp /tmp/adpanel-stop-XXXXXX.sh)
TMP_UPDATE=$(mktemp /tmp/adpanel-update-XXXXXX.sh)

echo "Скачивание: $START_URL"
wget -qO "$TMP_START" "$START_URL"
chmod +x "$TMP_START"
mv "$TMP_START" "$APP_DIR/start.sh"

echo "Скачивание: $STOP_URL"
wget -qO "$TMP_STOP" "$STOP_URL"
chmod +x "$TMP_STOP"
mv "$TMP_STOP" "$APP_DIR/stop.sh"

echo "Скачивание: $UPDATE_URL"
wget -qO "$TMP_UPDATE" "$UPDATE_URL"
chmod +x "$TMP_UPDATE"
mv "$TMP_UPDATE" "$APP_DIR/update.sh"

echo "Скрипты установлены: $APP_DIR/start.sh, $APP_DIR/stop.sh, $APP_DIR/update.sh"

# ============================
# 6.5. Создание .env файла
# ============================
echo "=== Создание .env файла ==="

ADMIN_LOGIN="admin"
ADMIN_PASSWORD=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 16 | head -n 1)

cat > "$APP_DIR/.env" <<EOF
ADMIN_LOGIN=$ADMIN_LOGIN
ADMIN_PASSWORD=$ADMIN_PASSWORD
EOF

# ============================
# 7. Создание systemd сервиса для сервера
# ============================
echo "=== Настройка systemd сервиса ==="
cat > /etc/systemd/system/adpanel-server.service <<EOF
[Unit]
Description=AdPanel Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$APP_DIR
ExecStart=$BINARY_PATH
Restart=on-failure
RestartSec=5
Environment=DATABASE_PATH=$APP_DIR/adpanel.db
Environment=PORT=$PORT

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable adpanel-server
systemctl start adpanel-server

# ============================
# 8. Настройка Nginx
# ============================
echo "=== Настройка Nginx ==="
NGINX_CONF="/etc/nginx/sites-available/adpanel"

cat > "$NGINX_CONF" <<EOF
server {
    listen $EXTERNAL_PORT;
    server_name $DOMAIN;

    # Проксирование API
    location /api/ {
        proxy_pass http://127.0.0.1:$PORT;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_cache_bypass \$http_upgrade;
    }

    # Проксирование других endpoint'ов сервера
    location /server {
        proxy_pass http://127.0.0.1:$PORT;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    # Статика клиента (по умолчанию)
    location / {
        root /var/www/adpanel;
        try_files \$uri \$uri/ /index.html;
    }
}
EOF

ln -sf "$NGINX_CONF" /etc/nginx/sites-enabled/adpanel
nginx -t && systemctl reload nginx

# ============================
# 9. Настройка фаервола
# ============================
echo "=== Настройка фаервола ==="
ufw allow $EXTERNAL_PORT/tcp
ufw allow 80/tcp
ufw allow 22/tcp
echo "y" | ufw enable 2>/dev/null || true

# ============================
# Готово
# ============================
echo ""
echo "========================================="
echo "  Установка завершена!"
echo "========================================="
echo ""
echo "Сервер: systemctl status adpanel-server"
echo "Nginx:  systemctl status nginx"
echo "Логи:   journalctl -u adpanel-server -f"
echo "Старт:  $APP_DIR/start.sh"
echo "Стоп:   $APP_DIR/stop.sh"
echo "Обнов:  $APP_DIR/update.sh"
echo ""
echo "Доступ: http://$(hostname -I | awk '{print $1}'):$EXTERNAL_PORT"
echo ""
echo ""
echo "========================================="
echo "  Учетные данные администратора"
echo "========================================="
echo "Логин:    $ADMIN_LOGIN"
echo "Пароль:   $ADMIN_PASSWORD"
echo "========================================="
