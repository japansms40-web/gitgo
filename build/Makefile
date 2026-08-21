APP := ghpublisher
LDFLAGS := -s -w
TAGS := migrated_fynedo

.PHONY: mac windows clean

# macOS 本机构建单文件
mac:
	CGO_ENABLED=1 go build -tags "$(TAGS)" -ldflags="$(LDFLAGS)" -o dist/$(APP) .

# Windows 交叉编译（需 x86_64-w64-mingw32-gcc，本机已装）
windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
		go build -tags "$(TAGS)" -ldflags="-H windowsgui $(LDFLAGS)" -o dist/$(APP).exe .

clean:
	rm -rf dist
