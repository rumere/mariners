BINARY_NAME=mpui

build:
	GOARCH=amd64 GOOS=linux go build -o ${BINARY_NAME} ui.go
	gtar cfz ${BINARY_NAME}.tar.gz ${BINARY_NAME} static/* tmpl/*

deploy:
	scp -i ~/.ssh/mplinksters.pem mpui.tar.gz ec2-user@s1.mplinksters.club:
	ssh -i ~/.ssh/mplinksters.pem ec2-user@s1.mplinksters.club "pkill mpui; rm -rf ui/*; cd ui; tar xfz ../mpui.tar.gz" 

start:
	ssh -i ~/.ssh/mplinksters.pem ec2-user@s1.mplinksters.club "cd ui; nohup ./mpui &"	

container:
	podman build -t ${BINARY_NAME} .

push:
	aws ecr get-login-password | podman login --username AWS --password-stdin 939932615330.dkr.ecr.us-west-1.amazonaws.com/${BINARY_NAME}
	podman push localhost/${BINARY_NAME}:latest 939932615330.dkr.ecr.us-west-1.amazonaws.com/${BINARY_NAME}

clean:
	go clean
	rm -f ${BINARY_NAME}
	rm -f ${BINARY_NAME}.tar.gz