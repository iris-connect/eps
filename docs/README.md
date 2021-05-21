# EPS Documentation

This repository builds the EPS documentation website.

## Setting up the environment

To build the docs, you will need a few tools installed on your machine:

- Node.js (for SASS compilation and possible transpiling of JS)
- Python 3, pip3 and virtualenv.
- `inotify-tools` or `fswatch` if you're on Mac (only required for the watch command).

You can then install all website dependencies by running

    make setup

This will set up a virtual Python environment, require the Python dependencies
into it and also install all required node modules.

## Building the docs

To build the docs simply run Make:

```bash
make
```

## Serving the docs

To serve the docs, simply run

```bash
make serve
```

This will start a HTTP server on port 8111 serving a hot-reloading version of
the docs.

## Fast Development With Auto-Reloading

To continuously update the build when things change, simply run

```bash
make watch
```

The watch command will automatically serve the docs as well using a
hot-reloading capable server.

## Translations

We use a machine-learning based translation service (DeepL) in combination with hand-tuning of inaccurate translations to generate different language-versions of this documentation. Our reference language for site content is `German`. To launch the auto-translation you'll need a valid DeepL-token. If you just want to contribute to the docs you can write your content in German or English and we'll take care of the translations. If you still want to experiment with the auto-translation mechanism you can simply run:

    TOKEN=[DeepL token] make translate

This will translate the configuration files, markdown texts and translation strings (used in HTML templates). You can fine-tune the translations by editing the corresponding YAML files ([filename].trans for configs and markdown). Be aware though that we use a hashing-based mechanism to detect outdated translations. Hence, if you modify a translated text and later modify the source text it will get re-translated and your modifications will be overwritten.

## Third-Party Code

We integrate the following third-party libraries directly into this codebase (in addition to the ones specified in `package.json` and `requirements.txt`):

* [Bulma](https://github.com/jgthms/bulma)
* [MathJax](https://github.com/mathjax/MathJax)
* [Open Sans](https://github.com/googlefonts/opensans)
* [Oxanium](https://github.com/sevmeyer/oxanium)
* [Source Code Pro](https://github.com/adobe-fonts/source-code-pro)
