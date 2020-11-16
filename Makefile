NAME = linkerd-disable-injection-mutation-webhook
IMAGE_PREFIX = ghcr.io/soluto
IMAGE_NAME = $(NAME)
IMAGE_VERSION = $$(git log --abbrev-commit --format=%h -s | head -n 1)

export GO111MODULE=on

app: deps
	go build -v -o $(NAME) ./...

deps:
	go get -v ./...

test: deps
	go test -v ./... -cover

docker:
	docker build --no-cache -t $(IMAGE_PREFIX)/$(IMAGE_NAME):$(IMAGE_VERSION) .
	docker tag $(IMAGE_PREFIX)/$(IMAGE_NAME):$(IMAGE_VERSION) $(IMAGE_PREFIX)/$(IMAGE_NAME):latest

push:
	@echo "WARNING: if you push to a public repo, you're pushing ssl key & cert, are you sure? [CTRL-C to cancel, ANY other to continue]"
	@sh read -n 1
	docker push $(IMAGE_PREFIX)/$(IMAGE_NAME):$(IMAGE_VERSION)
	docker push $(IMAGE_PREFIX)/$(IMAGE_NAME):latest

deploy:
	export KUBECONFIG=$$(kind get kubeconfig-path --name="kind"); kubectl apply -f deploy/

reset:
	export KUBECONFIG=$$(kind get kubeconfig-path --name="kind"); kubectl delete -f deploy/
	kind delete cluster --name kind

kind-deploy:
	docker build -t $(IMAGE_PREFIX)/$(IMAGE_NAME):latest .
	kind load docker-image $(IMAGE_PREFIX)/$(IMAGE_NAME):latest
	kubectl --context kind-kind delete pod -l app.kubernetes.io/instance=pull-secrets-mutation-webhook


.PHONY: docker push kind deploy reset kind-deploy
