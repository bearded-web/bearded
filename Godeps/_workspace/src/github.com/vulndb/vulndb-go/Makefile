.PHONY: all test lint vet fmt checkfmt prepare update-db

NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m
PKGSDIRS=$(shell find -L . -type f -name "*.go" -not -path "./Godeps/*")

all: test vet checkfmt

prepare: fmt test vet checkfmt

test:
	@echo "$(OK_COLOR)Test packages$(NO_COLOR)"
	@go test -v ./...

lint:
	@echo "$(OK_COLOR)Run lint$(NO_COLOR)"
	test -z "$$(golint ./... | grep -v Godeps/_workspace/src/ | tee /dev/stderr)"

vet:
	@echo "$(OK_COLOR)Run vet$(NO_COLOR)"
	@go vet ./...

checkfmt:
	@echo "$(OK_COLOR)Check formats$(NO_COLOR)"
	@./tools/checkfmt.sh .

fmt:
	@echo "$(OK_COLOR)Formatting$(NO_COLOR)"
	@echo $(PKGSDIRS) | xargs -I '{p}' -n1 goimports -w {p}

tools:
	@echo "$(OK_COLOR)Install tools$(NO_COLOR)"
	go get github.com/tools/godep
	go get golang.org/x/tools/cmd/goimports
	go get github.com/golang/lint
	go get github.com/jteeuwen/go-bindata/...

update-db:
	@echo "$(OK_COLOR)Update vulndb$(NO_COLOR)"
	@./tools/update-db.sh

