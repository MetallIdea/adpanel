#!/bin/bash
set -e

# ============================
# Конфигурация
# ============================
APP_DIR="/opt/adpanel"
BINARY_URL=""  # URL для скачивания дистрибутива сервера
DOMAIN=""      # Укажите домен или оставьте пустым
PORT=8080      # Порт приложения

# ============================
# 1. Обновление системы
# ============================
echo "=== Обновление системы ==="
apt update -y
apt upgrade -y

# ============================
# 2. Установка зависимостей
# ============================
echo "=== Установка зависимостей ==="
apt install -y \
    curl \
    wget \
    nginx \
    ufw

# ============================
# 3. Создание директории приложения
# ============================
echo "=== Подготовка директории ==="
mkdir -p "$APP_DIR"
mkdir -p /var/www/adpanel

# ============================
# 4. Скачивание и установка сервера
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
# 5. Создание systemd сервиса для сервера
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
# 6. Настройка Nginx
# ============================
echo "=== Настройка Nginx ==="
NGINX_CONF="/etc/nginx/sites-available/adpanel"

cat > "$NGINX_CONF" <<EOF
server {
    listen 80;
    server_name $DOMAIN;

    # Статика клиента
    location / {
        root /var/www/adpanel;
        try_files \$uri \$uri/ /index.html;
    }

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
    location / {
        proxy_pass http://127.0.0.1:$PORT;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF

ln -sf "$NGINX_CONF" /etc/nginx/sites-enabled/adpanel
nginx -t && systemctl reload nginx

# ============================
# 7. Настройка фаервола
# ============================
echo "=== Настройка фаервола ==="
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
echo ""
echo "Доступ: http://$(hostname -I | awk '{print $1}')"
echo ""
