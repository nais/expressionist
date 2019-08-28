Expressionist
=============

A simple Kubernetes admission webhook to validate `expr` in alert-resources. This is used together with [Alerterator](https://github.com/nais/alerterator/).

## Development

```
make setup-local # run once
make build local # after every update
```

### promtool

```
GO111MODULE=off go get -u github.com/prometheus/prometheus/cmd/promtool
```