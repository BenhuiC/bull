TAG_VERSION ?= latest
IMAGE ?= registry.sensetime.com/senseautoempower/marking:${TAG_VERSION}

.PHONY: docker_build
docker_build:
	docker build -t ${IMAGE} -f docker/Dockerfile .

docker_push: docker_build
	docker push ${IMAGE}

.PHONY: ci_build ci_deploy ci_release
ci_build: docker_push

ci_deploy:
	helm --kubeconfig ${KUBE_CONFIG} -n ${KUBE_NS} upgrade --install 	--values ${HELM_VALUES} \
	--set image.tag=${TAG_VERSION} \
	marking .deploy/chart/marking

ci_release:
	helm push --version ${CHART_VERSION} ${HELM_REPO} .deploy/chart/marking

.PHONY: build-agent
build-agent:
	go build -o .build/mark-agent agent/cmd/main.go
