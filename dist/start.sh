#!/bin/bash
set -e

echo "=== Запуск AdPanel Server ==="
systemctl daemon-reload
systemctl enable adpanel-server
systemctl start adpanel-server

echo "Ожидание запуска..."
sleep 2

if systemctl is-active --quiet adpanel-server; then
    echo "Сервер успешно запущен."
    echo "Статус: systemctl status adpanel-server"
    echo "Логи:   journalctl -u adpanel-server -f"
else
    echo "ОШИБКА: Сервер не запустился."
    journalctl -u adpanel-server -n 20
    exit 1
fi
