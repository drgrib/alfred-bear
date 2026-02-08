CMDS = search csearch create setcursor wiki-link link toc
WORKFLOW = alfred-bear.alfredworkflow
TMPDIR := $(shell mktemp -d)

.PHONY: all clean workflow

all: $(foreach cmd,$(CMDS),cmd/$(cmd)/$(cmd))

define BUILD_CMD
cmd/$(1)/$(1): cmd/$(1)/main.go $(wildcard core/*.go db/*.go)
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o $(TMPDIR)/$(1)-arm64 ./cmd/$(1)/
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 CC="clang -arch x86_64" go build -o $(TMPDIR)/$(1)-amd64 ./cmd/$(1)/
	lipo -create -output $$@ $(TMPDIR)/$(1)-arm64 $(TMPDIR)/$(1)-amd64
endef

$(foreach cmd,$(CMDS),$(eval $(call BUILD_CMD,$(cmd))))

workflow: all
	@rm -f $(WORKFLOW)
	zip -j $(WORKFLOW) info.plist icon.png
	$(foreach cmd,$(CMDS),zip $(WORKFLOW) cmd/$(cmd)/$(cmd);)
	@echo "Built $(WORKFLOW)"

clean:
	$(foreach cmd,$(CMDS),rm -f cmd/$(cmd)/$(cmd);)
	rm -f $(WORKFLOW)
