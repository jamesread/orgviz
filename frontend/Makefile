.PHONY: all deps build preview test lint prod

all: build

deps:
	npm install

build: deps
	npx vite build

prod: build

preview:
	npx vite preview

test:
	npm run test

lint:
	npm run lint
