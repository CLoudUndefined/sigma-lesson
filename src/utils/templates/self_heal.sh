#!/usr/bin/env bash
# На коленке, пять минут, простите. Взял backup.sh и cleanup.sh,
# слепил в один вечный цикл через nohup - до нормального автозапуска
# (cron, надо будет разобраться) руки пока не дошли.
#
# инцидент #14 за этот квартал, если что (веду счет с тех пор,
# как первый раз снес себе конфиг на коленке, в первую неделю тут)
#
# - Володя


ARCHIVE=/var/www/demo/backups
LOG=$ARCHIVE/health.log

while true; do
    sudo bash -c "echo '=== $(date) ===' >> $LOG" 2>/dev/null

    sudo chattr -R +i /var/www/demo 2>/dev/null

    if ! pgrep -x nginx > /dev/null; then
        sudo nginx -s reload 2>/dev/null || sudo nginx 2>/dev/null
    fi

    sudo tar -czf "$ARCHIVE/site_$(date +%s).tar.gz" /var/www/demo 2>/dev/null

    sudo bash -c "ls -t $ARCHIVE 2>/dev/null | tail -n +6 | xargs -I{} rm -f \"$ARCHIVE/{}\"" 2>/dev/null

    sleep 30
done
