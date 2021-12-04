BINARY_NAME=mpweather

build:
	GOARCH=amd64 GOOS=linux go build -o ${BINARY_NAME} weather.go

container:
	podman build -t ${BINARY_NAME} .

push:
	aws ecr get-login-password | podman login --username AWS --password-stdin 939932615330.dkr.ecr.us-west-1.amazonaws.com/${BINARY_NAME}
	podman push localhost/${BINARY_NAME}:latest 939932615330.dkr.ecr.us-west-1.amazonaws.com/${BINARY_NAME}

clean:
	go clean
	rm ${BINARY_NAME}