#!/bin/bash
set -e

# ============================
# Конфигурация
# ============================
APP_DIR="/opt/adpanel"
FRONTEND_URL="https://github.com/MetallIdea/adpanel/raw/refs/heads/main/dist/frontend.zip"
BINARY_URL="https://raw.githubusercontent.com/MetallIdea/adpanel/refs/heads/main/dist/server"
FRONTEND_DIR="/var/www/adpanel"

# ============================
# 1. Остановка сервиса
# ============================
echo "=== Остановка сервиса ==="
systemctl stop adpanel-server

if systemctl is-active --quiet adpanel-server; then
    echo "ОШИБКА: Не удалось остановить сервис."
    exit 1
fi
echo "Сервис остановлен."

# ============================
# 2. Обновление бинарника
# ============================
echo "=== Обновление бинарника ==="

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

echo "Бинарник обновлён: $BINARY_PATH"

# ============================
# 3. Обновление фронтенда
# ============================
echo "=== Обновление фронтенда ==="

if [ -n "$FRONTEND_URL" ]; then
    TMP_FRONTEND=$(mktemp /tmp/adpanel-frontend-XXXXXX.zip)
    echo "Скачивание фронтенда: $FRONTEND_URL"
    wget -qO "$TMP_FRONTEND" "$FRONTEND_URL"

    echo "Удаление старого фронтенда..."
    rm -rf "$FRONTEND_DIR"
    mkdir -p "$FRONTEND_DIR"

    echo "Распаковка нового фронтенда..."
    unzip -qo "$TMP_FRONTEND" -d "$FRONTEND_DIR"
    rm -f "$TMP_FRONTEND"

    echo "Фронтенд обновлён: $FRONTEND_DIR"
else
    echo "ПРЕДУПРЕЖДЕНИЕ: FRONTEND_URL не задан, фронтенд не обновлён"
fi

# ============================
# 4. Перезапуск сервиса
# ============================
echo "=== Перезапуск сервиса ==="
systemctl start adpanel-server

sleep 2

if systemctl is-active --quiet adpanel-server; then
    echo "Сервис успешно перезапущен."
    echo "Статус: systemctl status adpanel-server"
    echo "Логи:   journalctl -u adpanel-server -f"
else
    echo "ОШИБКА: Сервис не запустился."
    journalctl -u adpanel-server -n 20
    exit 1
fi

echo ""
echo "========================================="
echo "  Обновление завершено!"
echo "========================================="
