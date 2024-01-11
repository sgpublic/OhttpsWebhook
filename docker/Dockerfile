FROM docker:24.0.7-git

COPY ohttps-webhook /app/ohttps-webhook

RUN addgroup -S runner -g 1000 &&\
 addgroup -S runner-docker -g 997 &&\
 adduser -S runner -G runner -u 1000 &&\
 addgroup runner runner-docker &&\
 mkdir -p /app &&\
 chown -R runner:runner /app

USER runner

WORKDIR /app
ENTRYPOINT ["sh", "-c", "/app/ohttps-webhook"]
