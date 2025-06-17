FROM jlesage/baseimage:debian-12-v3

RUN apt-get update &&\
 apt-get install -y curl &&\
 curl -fsSL https://get.docker.com | bash &&\
 apt-get autoclean

COPY ./rootf /

ENV APP_NAME=ohttps-webhook
ARG APP_VERSION
ENV APP_VERSION=${APP_VERSION}
