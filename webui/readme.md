# Traefik Web UI

Access to Traefik Web UI, ex: http://localhost:8080

## Interface

Traefik Web UI provide 2 types of information:

- Providers with their backends and frontends information.
- Health of the web server.

## How to build (for backend developer)

Use the make file :

```shell
make build           # Generate Docker image
make generate-webui  # Generate static contents in `traefik/static/` folder.
```

## How to build (only for frontend developer)

- prerequisite: [Node 12.11+](https://nodejs.org) [Npm](https://www.npmjs.com/)

- Go to the directory `webui`

- To install dependencies, execute the following commands:

  - `npm install`

- Build static Web UI, execute the following command:

  - `npm run build`

- Static contents are build in the directory `static`

**Don't change manually the files in the directory `static`**

- The build allow to:
  - optimize all JavaScript
  - optimize all CSS
  - add vendor prefixes to CSS (cross-bowser support)
  - add a hash in the file names to prevent browser cache problems
  - all images will be optimized at build
  - bundle JavaScript in one file

## How to edit (only for frontend developer)

**Don't change manually the files in the directory `static`**

- Go to the directory `webui`
- Edit files in `webui/src`
- Run in development mode :
  - `npm run dev`

## Libraries

- [Node](https://nodejs.org)
- [Npm](https://www.npmjs.com/)
- [Webpack](https://github.com/webpack/webpack)
- [Vue](https://vuejs.org/)
- [Bulma](https://bulma.io)
- [D3](https://d3js.org)
- [D3 - Documentation](https://github.com/mbostock/d3/wiki)
