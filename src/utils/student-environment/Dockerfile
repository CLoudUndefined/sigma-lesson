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

COPY --from=signed-builder /build/signed /usr/local/bin/signed
RUN chmod 755 /usr/local/bin/signed

RUN mkdir -p /opt/student-environment/templates
COPY student-environment/bashrc-snippet.sh /opt/student-environment/bashrc-snippet.sh
COPY student-environment/templates/README-FIRST.txt /opt/student-environment/templates/README-FIRST.txt
COPY student-environment/templates/contract.txt /opt/student-environment/templates/contract.txt

COPY student-environment/supervisord.conf /etc/supervisor/conf.d/supervisord.conf

COPY student-environment/entrypoint.sh /opt/student-environment/entrypoint.sh
RUN chmod 755 /opt/student-environment/entrypoint.sh

EXPOSE 22 80

ENTRYPOINT ["/opt/student-environment/entrypoint.sh"]
