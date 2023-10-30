SERVICE_NAME=library
ENV=.env


HELM=helm/library-app

.PHONY: helm-run
helm-run:
	helm upgrade --install ${MY_RELEASE} ${HELM} -f ${HELM}/values.yaml  --set setter=lol \
		--namespace ${NAMESPACE} \
        --create-namespace \
        --atomic \
        --timeout 120s \
        --debug

.PHONY: helm-uninstall
helm-uninstall:
	helm uninstall ${MY_RELEASE} --namespace ${NAMESPACE}


.PHONY: helm-template
helm-template:
	helm template ${MY_RELEASE} ${HELM} --debug


.PHONY: helm-upgrade
helm-upgrade:
	helm upgrade ${MY_RELEASE} helm/library-app --namespace ${NAMESPACE}

.PHONY: k-clean
k-clean:
	kubectl delete sc,pvc,pv,cm,ing,secret,svc,all --all -n ${NAMESPACE}

.PHONY: helm-rollout
helm-rollout:
	helm rollback ${MY_RELEASE} --namespace ${NAMESPACE}


.PHONY: run
run:
	docker compose -f ./docker-compose.yaml --env-file $(ENV) up -d #--build #--remove-orphans

.PHONY: run-svc
run-svc: #  make run-svc svc=redis
	docker compose -f ./docker-compose.yaml --env-file $(ENV) up -d $(svc)

.PHONY: stop
stop:
	docker compose -f ./docker-compose.yaml --env-file $(ENV) down --volumes

.PHONY: remove-volume
remove-volume:
	docker volume rm lab2-template_db-data

#.PHONY: migrate-up
#migrate-up:
#	goose -dir "./migrations/sql/" postgres "host=localhost user=postgres password=postgres dbname=postgres sslmode=disable" up

#.PHONY: migrate-down
#migrate-down:
#	goose -dir "./migrations/sql/" postgres "host=localhost user=postgres password=postgres dbname=postgres sslmode=disable" down

.PHONY: lint
lint:
	go vet ./...
	golangci-lint run --fix

.PHONY: test
test:
	go test -v -race -timeout 90s -count=1 -shuffle=on  -coverprofile cover.out ./...
	@go tool cover -func cover.out | grep total | awk '{print $3}'
	go tool cover -html="cover.out" -o coverage.html

.PHONY: docker-login
docker-login:
	docker login -u ${REGISTRY_USER} -p ${REGISTRY_PASS}

.PHONY: image-build
image-build:
	docker compose -f ./docker-compose.yaml --env-file .env build

.PHONY: image-push
image-push:
	docker push ${IMAGE_NAME}:${IMAGE_TAG}


.PHONY: mocks
mocks:
	cd internal/handler; go generate;

.PHONY: .deps
deps: .deps
.deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.52.2
	go mod download


.PHONY: clean
clean:
	rm bin/${SERVICE_NAME}

.PHONY: clean-all
clean-all:
	sudo docker system prune --all --volumes -f

