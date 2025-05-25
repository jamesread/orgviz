default: deps prod

deps:
	npm install

prod:
	npx vite build --outDir ../frontend-dist --emptyOutDir

parcel:
	node_modules/.bin/parcel build src/index.pug --public-url "/webui/"
