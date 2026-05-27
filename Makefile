LAMBDA_FUNCTIONS := authorizer deleteDevice getAllDevices getDeviceByID saveDevice updateDevice

build:
	@echo "Building lambda binaries"
	@for func in $(LAMBDA_FUNCTIONS); do \
		env GOOS=linux GOARCH=arm64 go build -o build/lambda/$$func/bootstrap ./cmd/lambda/$$func/; \
	done

zip:
	@echo "Creating deployment packages"
	@for func in $(LAMBDA_FUNCTIONS); do \
		zip -j build/lambda/$$func.zip build/lambda/$$func/bootstrap; \
	done

all: clean build zip

clean:
	@echo "Cleaning up build directory"
	@rm -rf build/lambda

.PHONY: all build zip clean
