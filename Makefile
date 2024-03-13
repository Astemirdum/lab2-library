
SERVICE_NAME=library
ENV=.env

HELM=helm/library-app
NAMESPACE=default
MY_RELEASE=rsoi

.PHONY: helm-run-redpanda
helm-run-redpanda:
	#helm repo add redpanda https://charts.redpanda.com || true
	#helm repo update
	helm upgrade --install redpanda redpanda/redpanda \
        -n ${NAMESPACE} \
        --create-namespace \
        --set tls.enabled=false \
        --set external.domain=customredpandadomain.local \
        --set statefulset.initContainers.setDataDirOwnership.enabled=true \
        --set statefulset.replicas=1 \
        --set resources.cpu.cores=1 \
        --set resources.memory.max=300Mi \
        --set post_upgrade_job.enabled=false \
        --set post_install_job.enabled=false \
        --set storage.persistentVolume.size=1Gi

.PHONY: create-topics
create-topics:
	kubectl exec redpanda-0 -n ${NAMESPACE} -c redpanda -- rpk topic --brokers redpanda-0:9093 create library rating stats -p 1


.PHONY: helm-drop-redpanda
helm-drop-redpanda:
	helm uninstall redpanda -n ${NAMESPACE}

.PHONY: helm-run
helm-run:
	helm upgrade ${MY_RELEASE}-app ${HELM} -f ${HELM}/values.yaml  --set setter=lol --set setter1=lol1 \
		--install \
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
	helm upgrade ${MY_RELEASE} ${HELM} --namespace ${NAMESPACE}

.PHONY: helm-clean
helm-clean:
	kubectl delete sc,pvc,pv,cm,ing,secret,svc,all --all -n ${NAMESPACE}

.PHONY: helm-rollout
helm-rollout:
	helm rollback ${MY_RELEASE} --namespace ${NAMESPACE}


.PHONY: helm-db-run
helm-db-run:
	helm upgrade --install ${MY_RELEASE} \
		--set primary.initdb.scriptsConfigMap="postgresql-db-initdb-config" \
		--set primary.persistence.size="100Mi" \
 		oci://registry-1.docker.io/bitnamicharts/postgresql

.PHONY: run
run: #  make run svc=gateway
	docker compose -f ./docker-compose.yaml --env-file $(ENV) up -d --build $(svc)

.PHONY: stop
stop:
	docker compose -f ./docker-compose.yaml --env-file $(ENV) stop

.PHONY: down
down:
	docker compose -f ./docker-compose.yaml --env-file $(ENV) down volumes

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
	golangci-lint run --fix # --config .golangci.yml

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
	docker compose -f ./docker-compose.yaml --env-file .env build $(svc)


.PHONY: image-push # svc=gateway
image-push:
	docker push astdockerid1/$(svc):v1.0

SERVICES = gateway library provider stats library rating reservation
.PHONY: push-all-images
push-all-images:
	for service in $(SERVICES); do \
		docker push astdockerid1/$$service:v1.0; \
	done

.PHONY: image-clean
image-clean:
	docker rmi $(docker images -f "dangling=true" -q)

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

