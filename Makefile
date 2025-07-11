BINARY_NAME=vdlg-video-processor
AWS_EKS_CLUSTER_NAME=vdlg-eks-cluster

run:
	go run cmd/main.go

test-unit:
	go test -race -count=1 ./internal/... -coverprofile=coverage.out

test-e2e:
	go test -count=1 -timeout 300s ./test/...

test-coverage:
	go test -cover ./internal/... -coverpkg ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

pre-build:
	go mod download
	go mod verify
	go mod tidy

build:
	go build -o bin/${BINARY_NAME} -ldflags="-s -w" -tags appsec cmd/main.go

build-ci:
	go build -o bin/${BINARY_NAME} -ldflags="-s -w" -tags appsec cmd/main.go

build-docker:
	docker build -t tbtec/vdlg-video-processor:1.0.0 .	

run-docker:
	docker run -p 8080:8080 tbtec/vdlg-video-processor:1.0.0 --env-file .env

run-compose:
	docker compose up

run-compose-enviroment:
	docker compose -f docker-compose-enviroment.yaml up

docker-push:
	docker push tbtec/vdlg-video-processor:1.0.0

kube-eks-connect:
	aws eks update-kubeconfig --name ${AWS_EKS_CLUSTER_NAME} --region us-east-1

kube-config:
#	eval $(minikube docker-env)
	kubectl apply -f k8s/namespace.yaml

kube-deploy:
	kubectl apply -f k8s/configmap.yaml
	kubectl apply -f k8s/secret.yaml
	kubectl apply -f k8s/deployment.yaml
	kubectl apply -f k8s/service.yaml	
	kubectl apply -f k8s/ingress.yaml
	kubectl apply -f k8s/hpa.yaml

kube-deploy-eks:
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/configmap.yaml
#	kubectl apply -f k8s/secret.yaml
	kubectl apply -f k8s/deployment.yaml
	kubectl apply -f k8s/service.yaml
	kubectl apply -f k8s/ingress.yaml
	kubectl apply -f k8s/hpa.yaml
	
kube-deploy-eks-destroy:
	kubectl delete -f k8s/configmap.yaml
	kubectl delete -f k8s/secret.yaml
	kubectl delete -f k8s/deployment.yaml
	kubectl delete -f k8s/service.yaml
	kubectl delete -f k8s/ingress.yaml
	kubectl delete -f k8s/hpa.yaml
#	kubectl delete -f k8s/namespace.yaml