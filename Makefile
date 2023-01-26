# SPDX-FileCopyrightText: 2023 Kalle Fagerberg
#
# SPDX-License-Identifier: CC0-1.0

GO_FILES = $(shell git ls-files "*.go")

.PHONY: test
test:
	go test ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: deps
deps: deps-go deps-pip deps-npm

.PHONY: deps-go
deps-go:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
	go install github.com/mgechev/revive@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/yoheimuta/protolint/cmd/protolint@latest

.PHONY: deps-pip
deps-pip:
	python3 -m pip install --upgrade --user reuse

.PHONY: deps-npm
deps-npm: node_modules

node_modules:
	npm install

.PHONY: lint
lint: lint-md lint-go lint-license

.PHONY: lint-fix
lint-fix: lint-md-fix lint-go-fix

.PHONY: lint-md
lint-md: node_modules
	npx remark .

.PHONY: lint-md-fix
lint-md-fix: node_modules
	npx remark . -o

.PHONY: lint-go
lint-go:
	@echo goimports -d '**/*.go'
	@goimports -d $(GO_FILES)
	revive -formatter stylish -config revive.toml ./...

.PHONY: lint-go-fix
lint-fix-go:
	@echo goimports -d -w '**/*.go'
	@goimports -d -w $(GO_FILES)

.PHONY: lint-license
lint-license:
	reuse lint
