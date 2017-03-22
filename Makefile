
run: deps
	go env
	env | grep GITHUB
	go run *.go

deps:
	go get -u gopkg.in/yaml.v2
	go get -u github.com/google/go-github/github

.PHONY: run deps
