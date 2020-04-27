FROM golang:1.13-alpine as builder
RUN apk add --no-cache git
ENV GOOS=linux
ENV CGO_ENABLED=0
ENV GO111MODULE=on
COPY . /src
WORKDIR /src
RUN GOARCH=amd64 GO111MODULE=off go get -u github.com/prometheus/prometheus/cmd/promtool
RUN rm -f go.sum
RUN go get
RUN go test ./...
RUN go build -a -installsuffix cgo -o expressionist

FROM alpine:3.11
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /src/expressionist /app/expressionist
COPY --from=builder /go/bin/promtool /bin
EXPOSE 8080
EXPOSE 8443
CMD ["/app/expressionist"]
