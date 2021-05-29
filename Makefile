.DEFAULT_GOAL := run

version = `git fetch --tags >/dev/null && git describe --tags | cut -c 2-`
docker_container = akarasz/pajthy-backend
docker_tags = $(version),latest

docker_run = docker run --rm -i -t

.PHONY := docker
docker:
	docker build -t "$(docker_container):latest" .

.PHONY := run
run: docker
	docker_run -p 8000:8000 "$(docker_container):latest"

.PHONY := release
release: docker
	docker tag $(docker_container):latest $(docker_container):$(version)

.PHONY := push
push: release
	docker push $(docker_container):latest
	docker push $(docker_container):$(version)

# testing

.PHONY := test
test:
	go test ./...

.PHONY := test-long
test-long:
	go test -race -count=1 ./...