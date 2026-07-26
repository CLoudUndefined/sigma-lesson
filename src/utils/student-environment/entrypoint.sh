#!/usr/bin/env bash

set -e

if [ -z "$STUDENT_LOGIN" ] || [ -z "$STUDENT_PASSWORD" ]; then
    echo "Ошибка: STUDENT_LOGIN и STUDENT_PASSWORD должны быть заданы через docker run -e" >&2
    exit 1
fi

if ! id "$STUDENT_LOGIN" >/dev/null 2>&1; then
    echo "Создаю пользователя ${STUDENT_LOGIN}..."

    useradd -m -s /bin/bash "$STUDENT_LOGIN"
    echo "${STUDENT_LOGIN}:${STUDENT_PASSWORD}" | chpasswd
    usermod -aG sudo "$STUDENT_LOGIN"

    echo "${STUDENT_LOGIN} ALL=(ALL) NOPASSWD:ALL" > "/etc/sudoers.d/90-${STUDENT_LOGIN}"
    chmod 440 "/etc/sudoers.d/90-${STUDENT_LOGIN}"

    cat /opt/student-environment/bashrc-snippet.sh >> "/home/${STUDENT_LOGIN}/.bashrc"

    cp /opt/student-environment/templates/README-FIRST.txt "/home/${STUDENT_LOGIN}/README-FIRST.txt"
    cp /opt/student-environment/templates/contract.txt "/home/${STUDENT_LOGIN}/contract.txt"

    chown -R "${STUDENT_LOGIN}:${STUDENT_LOGIN}" "/home/${STUDENT_LOGIN}"

    echo "Пользователь ${STUDENT_LOGIN} создан."
else
    echo "Пользователь ${STUDENT_LOGIN} уже существует, пропускаю создание."
fi

exec /usr/bin/supervisord -c /etc/supervisor/conf.d/supervisord.conf
