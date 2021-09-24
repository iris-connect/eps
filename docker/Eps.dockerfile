FROM golang:1.16.3-alpine3.13 as builder
RUN apk add --update make && apk add --update openssl
RUN apk add --update bash
RUN bash --version
RUN bash
RUN mkdir -p /go/src /go/bin && chmod -R 777 /go
ENV GOPATH /go
ENV PATH /go/bin:$PATH
ARG VERSION
ENV VERSION=$VERSION
RUN echo $VERSION 
WORKDIR /app
COPY . .
RUN make

FROM alpine:latest

# Create a group and user
RUN addgroup --gid 9999 iris && adduser --disabled-password --gecos '' --uid 9999 -G iris -s /bin/ash iris

WORKDIR /app
COPY --from=builder /go/bin/eps /app/.scripts/entrypoint-eps.sh /app/

ENTRYPOINT ["/bin/sh", "./entrypoint-eps.sh"]
