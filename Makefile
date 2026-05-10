.PHONY: build-hooks test-hooks

HOOK_DEST := .claude/hooks/binaries
PRE_SRC   := hooks-logic/pre-tool-use
POST_SRC  := hooks-logic/post-tool-use
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

build-hooks:
	@mkdir -p $(HOOK_DEST)
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; arch=$${platform#*/}; \
		echo "Building pre-tool-use $$os/$$arch..."; \
		cd $(PRE_SRC) && CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
			go build -ldflags="-s -w" -o ../../$(HOOK_DEST)/pre-tool-use-$$os-$$arch . && cd ../..; \
		echo "Building post-tool-use $$os/$$arch..."; \
		cd $(POST_SRC) && CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
			go build -ldflags="-s -w" -o ../../$(HOOK_DEST)/post-tool-use-$$os-$$arch . && cd ../..; \
	done
	@echo "Binaries installed to $(HOOK_DEST)/"

test-hooks: build-hooks
	@cd $(PRE_SRC) && go test ./...
	@sh $(PRE_SRC)/test.sh
	@cd $(POST_SRC) && go test ./...
