default: testacc

.PHONY: testacc
testacc:
	TF_ACC=1 go test -race -parallel 3 ./... -v -timeout 5m
