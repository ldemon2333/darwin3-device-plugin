IMG ?= ldemon2333/darwin3-device-plugin:1.0

.PHONY: build-image 
build-image:
	docker build -t $(IMG) .