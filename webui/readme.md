# Traefik Web UI

Access to Traefik Web UI, ex: http://localhost:8080

## Interface

Traefik Web UI provide 2 types of information:

- Providers with their backends and frontends information.
- Health of the web server.

## How to build (for backend developer)

Use the make file :

```shell
make build-image                # Generate Docker image.
make clean-webui generate-webui # Generate static contents in `webui/static/` folder.
```

## How to build (only for frontend developer)

- prerequisite: [Node 12.11+](https://nodejs.org) [Yarn](https://yarnpkg.com/)

- Go to the `webui/` directory

- To install dependencies, execute the following commands:

  - `yarn install`

- Build static Web UI, execute the following command:

  - `yarn build`

- Static contents are built in the `webui/static/` directory

**Do not manually change the files in the `webui/static/` directory**

- The build allows to:
  - optimize all JavaScript
  - optimize all CSS
  - add vendor prefixes to CSS (cross-browser support)
  - add a hash in the file names to prevent browser cache problems
  - optimize all images at build time
  - bundle JavaScript in one file

## How to edit (only for frontend developer)

**Do not manually change the files in the `webui/static/` directory**

- Go to the `webui/` directory
- Edit files in `webui/src/`
- Run in development mode :
  - `yarn dev`

## Libraries

- [Node](https://nodejs.org)
- [Yarn](https://yarnpkg.com/)
- [Webpack](https://github.com/webpack/webpack)
- [Vue](https://vuejs.org/)
- [Bulma](https://bulma.io)
- [D3](https://d3js.org)
- [D3 - Documentation](https://github.com/mbostock/d3/wiki)
