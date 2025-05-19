# Traefik Web UI

Access to Traefik Web UI, ex: http://localhost:8080

## Interface

Traefik Web UI provide 2 types of information:

- Providers with their backends and frontends information.
- Health of the web server.

## How to build (for backend developer)

Use the Makefile :

```shell
make build-image                # Generate Docker image.
make clean-webui generate-webui # Generate static contents in `webui/static/` folder.
```

## How to build (only for frontend developer)

- prerequisite: [Node 22](https://nodejs.org) [Yarn](https://yarnpkg.com/)

- Go to the `webui/` directory

- As we use Yarn v4, you will need to enable corepack before installing dependencies:

  - `corepack enable`

- To install dependencies, execute the following commands:

  - `yarn install`

- Build static Web UI, execute the following command:

  - `yarn build`

- Static contents are built in the `webui/static/` directory

**Do not manually change the files in the `webui/static/` directory**

The build allows to:
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
- Create and populate the `.env` file using the values inside `.env.sample` file.
- Run in development mode :
  - `yarn dev`
- The application will be available at `http://localhost:3000/`. On development mode, the application will run with mocked data served by [Mock Service Worker](https://mswjs.io/).

## How to run tests

- Execute the following commands:
  - `yarn test`
  - or `yarn test:watch` if you want them in watch mode

## Libraries

- [Node](https://nodejs.org)
- [Yarn](https://yarnpkg.com/)
- [React](https://reactjs.org/)
- [Vite](https://vitejs.dev/)
- [Faency](https://github.com/containous/faency)
- [Vitest](https://vitest.dev/)
- [Mock Service Worker](https://mswjs.io/)
