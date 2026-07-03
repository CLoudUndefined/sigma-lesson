FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    nginx \
    openssh-server \
    inotify-tools \
    curl \
    wget \
    nano \
    vim \
    less \
    coreutils \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir /var/run/sshd
RUN echo 'PermitRootLogin no' >> /etc/ssh/sshd_config
RUN echo 'PasswordAuthentication yes' >> /etc/ssh/sshd_config

RUN useradd -m -s /bin/bash student && echo "student:sigma2026" | chpasswd

RUN mkdir -p /etc/nginx/sites-enabled && \
    chown root:student /etc/nginx/sites-enabled && \
    chmod 775 /etc/nginx/sites-enabled

COPY setup.sh /setup.sh
COPY entrypoint.sh /entrypoint.sh
COPY watch-nginx.sh /watch-nginx.sh
RUN chmod +x /setup.sh /entrypoint.sh /watch-nginx.sh

EXPOSE 22 80

ENTRYPOINT ["/entrypoint.sh"]
