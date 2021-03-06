DATE=$(shell date "+%Y-%m-%d")
LAST_COMMIT=$(shell git --no-pager log -1 --pretty=%h)
VERSION="$(DATE)-$(LAST_COMMIT)"
LDFLAGS := -X github.com/nais/expressionist/pkg/version.Revision=$(shell git rev-parse --short HEAD) -X github.com/nais/expressionist/pkg/version.Version=$(VERSION)

build:
	go build

test:
	go test ./... -count=1

release:
	go build -a -installsuffix cgo -o expressionist -ldflags "-s $(LDFLAGS)"

setup-local:
	kubectl apply -f webhook.yaml
	./gen-cert.sh
	./ca-bundle.sh

local:
	./expressionist --cert ./cert.pem --key ./key.pem
