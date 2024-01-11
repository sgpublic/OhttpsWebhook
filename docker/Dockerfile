FROM docker:24.0.7-git

RUN useradd -m -u 1000 runner &&\
 usermod -aG docker runner &&\
 mkdir -p /app

COPY ohttps-webhook /app/ohttps-webhook
RUN chown -R runner:runner /app

USER runner

WORKDIR /app
ENTRYPOINT ["/app/ohttps-webhook"]
