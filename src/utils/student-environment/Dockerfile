# syntax=docker/dockerfile:1

FROM golang:1.26-bookworm AS signed-builder

WORKDIR /build
COPY signed/go.mod ./
COPY signed/main.go ./
COPY signed/Makefile ./

RUN make build

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    openssh-server \
    nginx \
    sudo \
    tree \
    e2fsprogs \
    nano \
    cron \
    supervisor \
    ncurses-term \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /var/run/sshd /var/log/supervisor

ARG STUDENT_LOGIN=student
ARG STUDENT_PASSWORD=changeme

RUN useradd -m -s /bin/bash "${STUDENT_LOGIN}" \
    && echo "${STUDENT_LOGIN}:${STUDENT_PASSWORD}" | chpasswd \
    && usermod -aG sudo "${STUDENT_LOGIN}" \
    && echo "${STUDENT_LOGIN} ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/90-${STUDENT_LOGIN} \
    && chmod 440 /etc/sudoers.d/90-${STUDENT_LOGIN}

COPY --from=signed-builder /build/signed /usr/local/bin/signed
RUN chmod 755 /usr/local/bin/signed

COPY student-environment/bashrc-snippet.sh /tmp/bashrc-snippet.sh
RUN cat /tmp/bashrc-snippet.sh >> /home/${STUDENT_LOGIN}/.bashrc \
    && rm /tmp/bashrc-snippet.sh

COPY student-environment/templates/README-FIRST.txt /home/${STUDENT_LOGIN}/README-FIRST.txt
COPY student-environment/templates/contract.txt /home/${STUDENT_LOGIN}/contract.txt

RUN chown -R ${STUDENT_LOGIN}:${STUDENT_LOGIN} /home/${STUDENT_LOGIN}

COPY student-environment/supervisord.conf /etc/supervisor/conf.d/supervisord.conf

EXPOSE 22 80

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
