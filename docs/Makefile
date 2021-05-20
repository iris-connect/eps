SITE = "src/site-dev.yml"
SHELL := /bin/bash
RELEASE_HOST ?= https://raw.githubusercontent.com/adewes/

all:	site

export PATH := $(PATH):$(HOME)/.local/bin

setup: virtualenv requirements

virtualenv:
	virtualenv --python python3 venv

requirements:
	venv/bin/pip install -r requirements.txt
	npm ci

sass:
	node_modules/.bin/node-sass src/assets/scss/main.scss src/static/css/main.css

babel:
	node_modules/.bin/babel src/**/*.es6 --out-dir "." || true

translate:
	@if [ -n "$(TOKEN)" ]; then venv/bin/beam i18n translate $(TOKEN) src $(TA); fi;

translate-config:
	@if [ ! -n "$(TOKEN)" ]; then echo "Please set the translation token as the TOKEN variable." && exit 1; fi;
	venv/bin/beam i18n translate-config $(TOKEN) src $(TA)

downloads:
	curl $(RELEASE_HOST)zilpzalp/master/apps/releases.yml > src/en/releases.yml

optimize: optimize-images optimize-html

optimize-images:
	find src -name "*.png" -exec optipng {} \;

optimize-html:
	find build -name "*.html" -exec tidy -m {} \;

site: translate sass babel
	venv/bin/beam -vv up --site $(SITE)

clean:
	rm -rf build/*

serve:
	python3 -m http.server -d build 8112

watch: site
	./watch.sh


