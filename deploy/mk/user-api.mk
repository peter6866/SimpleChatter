VERSION=latest

SERVER_NAME=user
SERVER_TYPE=api

# docker repo
DOCKER_REPO_TEST=637423626125.dkr.ecr.us-east-2.amazonaws.com/simplechatter/${SERVER_NAME}-${SERVER_TYPE}-dev
# docker repo tag
VERSION_TEST=$(VERSION)
# docker repo name
APP_NAME_TEST=simplechatter-${SERVER_NAME}-${SERVER_TYPE}-test

# dockerfile
DOCKER_FILE_TEST=./deploy/dockerfile/Dockerfile_${SERVER_NAME}_${SERVER_TYPE}_dev

# docker build
build-test:

	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/${SERVER_NAME}-${SERVER_TYPE} ./apps/${SERVER_NAME}/${SERVER_TYPE}/${SERVER_NAME}.go
	docker build . -f ${DOCKER_FILE_TEST} --no-cache -t ${APP_NAME_TEST}

# docker tag
tag-test:

	@echo 'create tag ${VERSION_TEST}'
	docker tag ${APP_NAME_TEST} ${DOCKER_REPO_TEST}:${VERSION_TEST}

publish-test:

	@echo 'publish ${VERSION_TEST} to ${DOCKER_REPO_TEST}'
	docker push $(DOCKER_REPO_TEST):${VERSION_TEST}

release-test: build-test tag-test publish-test
