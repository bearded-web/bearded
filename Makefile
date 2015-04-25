.PHONY: all test lint vet fmt travis coverage


all: test vet fmt


travis: test fmt vet coverage


test:
	@echo "+ $@"
	go test -v -cover ./...

coverage:
	@echo "+ $@"
	@./coverage.sh
	-goveralls -coverprofile=gover.coverprofile -service=travis-ci

lint:
	@echo "+ $@"
	test -z "$$(golint ./... | grep -v Godeps/_workspace/src/ | tee /dev/stderr)"


vet:
	@echo "+ $@"
	go vet ./...


fmt:
	@echo "+ $@"
	./checkfmt.sh .

updep:
	@echo "+ $@"

	GOOS=linux godep save ./...