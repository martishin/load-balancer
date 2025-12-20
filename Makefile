.PHONY: test

test:
	go test ./...

build:
	docker build -t load-balancer .

run:
	docker-compose up -d

stop:
	docker-compose stop

stop-web:
	docker-compose stop web1 web2

remove:
	docker-compose down

logs:
	docker-compose logs -f

KUBE_NS := load-balancer
DEPLOYMENT_NAME := load-balancer

k8s-ns:
	kubectl create namespace $(KUBE_NS) --dry-run=client -o yaml | kubectl apply -f -

k8s-apply: k8s-ns
	kubectl -n $(KUBE_NS) apply -f k8s/web.yaml
	kubectl -n $(KUBE_NS) apply -f k8s/load-balancer.yaml
	kubectl -n $(KUBE_NS) apply -f k8s/ingress.yaml
	$(MAKE) k8s-wait

k8s-deploy: k8s-apply

k8s-release:
	kubectl -n $(KUBE_NS) rollout restart deploy/$(DEPLOYMENT_NAME)
	$(MAKE) k8s-wait

k8s-wait:
	kubectl -n $(KUBE_NS) rollout status deploy/web --timeout=120s
	kubectl -n $(KUBE_NS) rollout status deploy/$(DEPLOYMENT_NAME) --timeout=120s

k8s-status:
	kubectl -n $(KUBE_NS) get all
	kubectl -n $(KUBE_NS) get ingress,certificate

k8s-describe:
	kubectl -n $(KUBE_NS) describe deploy/$(DEPLOYMENT_NAME)
	kubectl -n $(KUBE_NS) describe ingress/load-balancer

k8s-clean:
	kubectl delete namespace $(KUBE_NS) --ignore-not-found
