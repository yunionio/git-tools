ROOT_DIR = $(CURDIR)
OUTPUT_DIR = $(ROOT_DIR)/_output
BIN_DIR = $(OUTPUT_DIR)/bin
REPO_PREFIX = yunion.io/x/git-tools

GO_BUILD := go build

CMDS := $(shell find $(ROOT_DIR)/cmd -mindepth 1 -maxdepth 1 -type d)

VERSION ?= $(shell git describe --exact-match 2> /dev/null || \
	                   git describe --match=$(git rev-parse --short=8 HEAD) --always --dirty --abbrev=8)

build: clean
	@for CMD in $(CMDS); do \
		echo build $$CMD; \
		$(GO_BUILD) -o $(BIN_DIR)/`basename $${CMD}` $$CMD; \
	done

prepare_dir:
	@mkdir -p $(BIN_DIR)

test:
	go test -v ./...

cmd/%: prepare_dir
	$(GO_BUILD) -o $(BIN_DIR)/$(shell basename $@) $(REPO_PREFIX)/$@

clean:
	@rm -rf $(BIN_DIR)
	@rm -rf $(OUTPUT_DIR)/changelog/release-*

%:
	@:

.PHONY: build clean
