FROM alpine:latest
RUN apk add --update make && apk add --update openssl
RUN apk add --update bash
RUN apk add --update coreutils && rm -rf /var/cache/apk/*
RUN bash --version
RUN bash

WORKDIR /app
COPY . .
ENTRYPOINT [ "make" ]
