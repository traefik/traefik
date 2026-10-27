# Documentation

## Tooling

| Tool              | Documentation                       | Sources                           |
|-------------------|-------------------------------------|-----------------------------------|
| mkdocs            | [documentation][mkdocs]             | [Sources][mkdocs-src]             |
| mkdocs-material   | [documentation][mkdocs-material]    | [Sources][mkdocs-material-src]    |
| pymdown-extensions| [documentation][pymdown-extensions] | [Sources][pymdown-extensions-src] |

[mkdocs]: https://www.mkdocs.org "Mkdocs"
[mkdocs-src]: https://github.com/mkdocs/mkdocs "Mkdocs - Sources"

[mkdocs-material]: https://squidfunk.github.io/mkdocs-material/ "Material for MkDocs"
[mkdocs-material-src]: https://github.com/squidfunk/mkdocs-material "Material for MkDocs - Sources"

[pymdown-extensions]: https://facelessuser.github.io/pymdown-extensions "PyMdown Extensions"
[pymdown-extensions-src]: https://github.com/facelessuser/pymdown-extensions "PyMdown Extensions - Sources"

## Build locally without docker

```sh
# Pre-requisite: python3, pip and virtualenv
DOCS="/tmp/traefik-docs"
mkdir "$DOCS"
virtualenv "$DOCS"
source "$DOCS/bin/activate"
pip install -r requirements.txt
mkdocs serve # or mkdocs build
```
