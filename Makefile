build:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -o bin/admission-webhook-server
image-build: build
	docker build -t registry.cn-shanghai.aliyuncs.com/carl-zyc/admission-webhook-server:v1 .