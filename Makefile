default: proto service frontend

proto:
	$(MAKE) -wC proto

service:
	$(MAKE) -wC service

service-build:
	$(MAKE) -wC service build

service-run:
	$(MAKE) -wC service run

service-test:
	$(MAKE) -wC service test

service-lint:
	$(MAKE) -wC service lint

frontend:
	$(MAKE) -wC frontend

frontend-build:
	$(MAKE) -wC frontend build

frontend-test:
	$(MAKE) -wC frontend test

frontend-lint:
	$(MAKE) -wC frontend lint

docs:
	$(MAKE) -wC docs
	./docs/node_modules/.bin/antora antora-playbook.yml

lint: service-lint frontend-lint

test: service-test frontend-test

integration-test: service-build frontend-build
	$(MAKE) -wC integration-tests

it: integration-test

screenshots:
	$(MAKE) -wC var/marketing/screenshots

docker: container
container:
	buildah bud

.PHONY: default proto service service-build service-run service-test service-lint frontend frontend-build frontend-test frontend-lint docs lint test integration-test it screenshots docker container
