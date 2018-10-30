# Traefik Web UI

Access to Traefik Web UI, ex: http://localhost:8080

## Interface

Traefik Web UI provide 2 types of informations:
- Providers with their backends and frontends information.
- Health of the web server.

## How to build (for backends developer)

Use the make file :

```shell
make build           # Generate Docker image
make generate-webui  # Generate static contents in `traefik/static/` folder.
```

## How to build (only for frontends developer)

- prerequisite: [Node 6+](https://nodejs.org) [yarn](https://yarnpkg.com/)

  Note: In case of conflict with the Apache Hadoop Yarn Command Line Interface, use the `yarnpkg`
  alias.

- Go to the directory `webui`

- To install dependencies, execute the following commands:
  - `yarn install`

- Build static Web UI, execute the following command:
  - `yarn run build`

- Static contents are build in the directory `static`

**Don't change manually the files in the directory `static`**

- The build allow to:
  - optimize all JavaScript
  - optimize all CSS
  - add vendor prefixes to CSS (cross-bowser support)
  - add a hash in the file names to prevent browser cache problems
  - all images will be optimized at build
  - bundle JavaScript in one file


## How to edit (only for frontends developer)

**Don't change manually the files in the directory `static`**

- Go to the directory `webui`
- Edit files in `webui/src`
- Run in development mode :
  - `yarn start`

## Libraries

- [Node](https://nodejs.org)
- [Yarn](https://yarnpkg.com/)
- [Webpack](https://github.com/webpack/webpack)
- [Angular](https://angular.io)
- [Bulma](https://bulma.io)
- [D3](https://d3js.org)
- [D3 - Documentation](https://github.com/mbostock/d3/wiki)
