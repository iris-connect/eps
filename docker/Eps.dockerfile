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
WORKDIR /app
COPY --from=builder /go/bin/eps /app/eps
ENTRYPOINT ["./eps"]
