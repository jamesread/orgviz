default: proto service frontend

proto:
	$(MAKE) -wC proto

service:
	$(MAKE) -wC service

frontend:
	$(MAKE) -wC frontend

py:
	./orgviz.py

docker: container
container:
	buildah bud

docs:
	$(MAKE) -wC docs
	./docs/node_modules/.bin/antora antora-playbook.yml
#	./orgviz.py --profilePictures --profilePictureDirectory examples/profilePics/ -T png -I examples/ExampleCompany.org --dpi 300
#	mv orgviz.png docs/ExampleCompany.png
#	asciidoctor README.adoc

tests: test
test:
	coverage run --branch --source orgviz -m pytest
	coverage report
	coverage html

tests-debian:
	python3-coverage run --branch --source orgviz -m pytest
	python3-coverage report
	python3-coverage html

lint:
	pylint-3 orgviz

lint-debian:
	pylint orgviz

.PHONY: docs default test default service proto frontend
