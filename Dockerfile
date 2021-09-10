FROM prom/prometheus:latest as promtool

FROM golang:1.17-alpine as builder
ENV GOOS=linux
ENV CGO_ENABLED=0
COPY --from=promtool /bin/promtool /bin

WORKDIR /workspace
COPY . .
RUN go get
RUN go test ./...
RUN go build -a -installsuffix cgo -o expressionist

FROM alpine:3.14
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /workspace/expressionist /app/expressionist
COPY --from=promtool /bin/promtool /bin
EXPOSE 8080
EXPOSE 8443
CMD ["/app/expressionist"]
