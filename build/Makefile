.PHONY: mac windows clean

# macOS 本机构建（wails build 会先编译前端再打包成 .app）
mac:
	wails build -platform darwin/arm64

# Windows 交叉编译
windows:
	wails build -platform windows/amd64

clean:
	rm -rf build/bin
