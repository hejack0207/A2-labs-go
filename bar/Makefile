.PHONY: build image container deploy undeploy dinfo

KUBECTL-OPTS =
KUBDIR = k8s
KUB = sudo kubectl $(KUBECTL-OPTS)
NS = --namespace foo

prepare-env:
	#install k3s before all steps below
	curl -sfL https://get.k3s.io | sh -

build:
	go build -o docker/bar app/bar.go

image:
	docker build -t bar:0.0.1 docker
	sudo mkdir -p /opt/{registry,certs}
	grep /etc/hosts '127.0.0.1 dev' || echo '127.0.0.1 dev' | sudo tee -a /etc/hosts
	# during key generation, when prompting to enter hostname/servername,
	# please MUST enter `dev` here, otherwise the hostname will not match
	sudo openssl req -newkey rsa:4096 -nodes -sha256 -keyout /opt/certs/registry.key -x509 -days 365 -out /opt/certs/registry.crt
	sudo cp /opt/certs/registry.crt /etc/pki/ca-trust/source/anchors/
	sudo update-ca-trust
	sudo systemctl restart docker
	docker tag bar:0.0.1 dev:31320/bar:0.0.1

deploy:
	$(KUB) apply -f $(KUBDIR)/private-registry.yaml
	$(KUB) apply -f $(KUBDIR)/private-registry-svc.yaml
	sleep 5s
	docker push dev:31320/bar:0.0.1
	$(KUB) get namespace foo || $(KUB) create namespace foo
	$(KUB) apply -f $(KUBDIR)/bar-deployment.yaml --record

deploy-ex:
	$(KUB) $(NS) expose deployment deployment-bar --type=LoadBalancer --name=service-bar

undeploy:
	sudo rm -rf /opt/registry/*
	sudo rm -rf /opt/certs/*
	-$(KUB) delete deployment private-repository-k8s
	-$(KUB) delete service private-repository-k8s
	-$(KUB) $(NS) delete deployment deployment-bar

dinfo:
	-$(KUB) get deployments/private-repository-k8s
	-$(KUB) get service/private-repository-k8s
	-$(KUB) $(NS) get deployments/deployment-bar

testapp:
	-@docker kill bar-0.0.1 &>/dev/null
	-@docker rm bar-0.0.1 &>/dev/null
	@docker run -d --name bar-0.0.1 -p 8080:8080 bar:0.0.1 &>/dev/null
	@sleep 5s
	@[[ $$(curl -sL http://localhost:8080/welcome?name="abc") = "Welcome abc" ]] && echo "bar is WORKING" || echo "bar is not WORKING"

netpolicy:
	$(KUB) apply -f $(KUBDIR)/bar-netpolicy.yaml

ainfo:
	$(KUB) get pods --all-namespaces -o wide
	$(KUB) get deployments --all-namespaces
	$(KUB) get services --all-namespaces

package:
	-rm ../SRE-code-exercise-bar-byhelong.zip
	zip -r ../SRE-code-exercise-bar-byhelong.zip . -x .git/\* bar



