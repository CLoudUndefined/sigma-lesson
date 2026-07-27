#!/bin/bash
# копируй строки по одной, разбираясь, что делает каждая.

# > перезаписывает файл целиком - с этой строки дашборд начинается заново
# при каждом запуске скрипта.
echo "=== СТАТУС НА $(date) ===" > /var/www/demo/dashboard.txt

# Дальше используем >> - дописываем в конец файла, не стирая то, что
# уже написали строкой выше.
echo "" >> /var/www/demo/dashboard.txt

echo "Место на диске:" >> /var/www/demo/dashboard.txt
df -h / >> /var/www/demo/dashboard.txt
echo "" >> /var/www/demo/dashboard.txt

echo "Оперативная память:" >> /var/www/demo/dashboard.txt
free -h >> /var/www/demo/dashboard.txt
echo "" >> /var/www/demo/dashboard.txt

echo "Последние бэкапы:" >> /var/www/demo/dashboard.txt
ls -l ~/backups | tail -n 3 >> /var/www/demo/dashboard.txt

echo "" >> /var/www/demo/dashboard.txt
echo "Отчет готов. Все свободны" >> /var/www/demo/dashboard.txt
