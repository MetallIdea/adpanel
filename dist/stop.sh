#!/bin/bash

echo "=== Остановка AdPanel Server ==="
systemctl stop adpanel-server

if systemctl is-active --quiet adpanel-server; then
    echo "ОШИБКА: Сервер не удалось остановить."
    exit 1
else
    echo "Сервер успешно остановлен."
fi
