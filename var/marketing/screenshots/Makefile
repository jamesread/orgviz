# Regenerates marketing screenshots against a throwaway orgviz instance
# seeded with fictional org files. Run `make` here; the process is always
# stopped afterwards.
#
# orgviz has no database or broker: charts are YAML files read at start,
# and avatars are static files.

ROOT := $(abspath ../../..)
HERE := $(abspath .)
TMP := $(HERE)/tmp

BACKEND_PORT := 18480
BACKEND_URL := http://127.0.0.1:$(BACKEND_PORT)

DOCS_IMAGES := $(ROOT)/docs/modules/ROOT/images/marketing
DOCS_PNGS := home.png second-company.png home-dark.png

.PHONY: all build up seed shoot manifest publish down clean

all:
	@$(MAKE) build up seed shoot manifest publish; rc=$$?; $(MAKE) down; exit $$rc

build:
	$(MAKE) -C $(ROOT)/service build
	[ -d $(ROOT)/frontend/node_modules ] || (cd $(ROOT)/frontend && npm ci)
	$(MAKE) -C $(ROOT)/frontend build

up:
	mkdir -p $(TMP)
	cd $(ROOT)/service && { ORGVIZ_LISTEN=127.0.0.1:$(BACKEND_PORT) \
		ORGVIZ_FRONTEND_DIR=$(ROOT)/frontend/dist \
		$(ROOT)/service/orgviz -configdir $(HERE)/seed \
		> $(TMP)/service.log 2>&1 & echo $$! > $(TMP)/service.pid; }
	@for i in $$(seq 60); do curl -fsS -X POST -H 'Content-Type: application/json' -d '{}' \
		$(BACKEND_URL)/api/clientapi.ClientToFrontend/getClientInitialSettings >/dev/null 2>&1 && exit 0; \
		sleep 1; done; \
		echo "Backend did not become ready; see $(TMP)/service.log"; exit 1

seed:
	./seed.sh $(BACKEND_URL)

shoot:
	repo-helper screenshot --config screenshots.ini

manifest:
	./generate-manifest.py

publish:
	mkdir -p $(DOCS_IMAGES)
	cp $(DOCS_PNGS) $(DOCS_IMAGES)/

down:
	-@[ -f $(TMP)/service.pid ] && kill $$(cat $(TMP)/service.pid) 2>/dev/null; rm -f $(TMP)/service.pid; true

clean: down
	rm -rf $(TMP)
