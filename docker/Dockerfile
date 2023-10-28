FROM debian:bookworm-20230904-slim

RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/' /etc/apt/sources.list.d/*.sources &&\
 useradd -m -u 1000 runner &&\
 curl -fsSL https://get.docker.com | bash &&\
 usermod -aG docker runner &&\
 systemctl disable docker --now &&\
 mkdir -p /app

COPY ohttps-webhook /app/ohttps-webhook
RUN chown -R runner:runner /app

USER runner

WORKDIR /app
ENTRYPOINT ["/app/ohttps-webhook"]
