# TODO: This can be taken from go.mod
app_name=word-of-wisdom

local_build_registry:=localhost:6000/$(app_name)-build
local_client_registry:=localhost:6000/$(app_name)-client
local_server_registry:=localhost:6000/$(app_name)-server

docker_build_args=\
		--build-arg BASE_BUILD_IMAGE=$(shell cat ./.build-image) \
		-f Dockerfile.build \
		-t $(local_build_registry):latest \
		../

.local-build-image:
	docker build $(docker_build_args)
	docker inspect $(local_build_registry):latest > $@-inspect
	cat $@-inspect | jq -r '.[0].RepoTags[] | select(. | startswith("$(local_build_registry)"))' > $@


docker_out_args=\
		--build-arg BUILD_IMAGE=$(shell cat .local-build-image) \
		--build-arg RUNTIME_IMAGE=$(shell cat .runtime-image) \
		-f Dockerfile.out \
		../

.local-client-image: .local-build-image
	docker build --build-arg TARGET=client -t $(local_client_registry):latest $(docker_out_args)
	docker inspect $(local_client_registry):latest > $@-inspect
	cat $@-inspect | jq -r '.[0].RepoTags[] | select(. | startswith("$(local_client_registry)"))' > $@

.local-server-image: .local-build-image
	docker build --build-arg TARGET=server -t $(local_server_registry):latest $(docker_out_args)
	docker inspect $(local_server_registry):latest > $@-inspect
	cat $@-inspect | jq -r '.[0].RepoTags[] | select(. | startswith("$(local_server_registry)"))' > $@

.PHONY: clean-images
clean-images:
	rm -f .local-*