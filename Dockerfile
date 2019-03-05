FROM golang:1.12-alpine as builder
RUN apk add --no-cache git make
ENV GOOS=linux
ENV CGO_ENABLED=0
ENV GO111MODULE=on
WORKDIR /root/src

COPY Makefile /root/src
COPY go.mod /root/src
COPY main.go /root/src
COPY pkg/ /root/src/pkg/

RUN go get
RUN go test ./...
RUN make release

FROM alpine:3.9
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /root/src/expressionist /app/expressionist
EXPOSE 8080
EXPOSE 8443
CMD ["/app/expressionist"]
