# Traefik Proxy dashboard

Documentation related to the agent dashboard of Traefik Proxy.

## Interface

Traefik Proxy dashboard provide information about HTTP, TCP, and UDP resources.

## How to build (for backend developers)

Use the Makefile:

```shell
make build           # Generate Docker image
```

## How to build (for frontend developers)

- prerequisite: [Node Stable 18+](https://nodejs.org)

- To install dependencies, run `yarn`

- To build static Web UI, run `yarn build`

- Static contents are built in the `dist` directory

**Don't manually change the files in the directory `dist`**

- The build allows to:
  - optimize all JavaScript
  - optimize all CSS
  - add vendor prefixes to CSS (cross-bowser support)
  - add a hash in the file names to prevent browser cache problems
  - all images will be optimized at build
  - bundle JavaScript into one file

## How to edit (only for frontend developers)

- Run in development mode
  - `yarn dev`

## How to run tests

- Execute the following commands:
  - `yarn test`
  - or `yarn test:watch` if you want them in watch mode

## Libraries

- [Node](https://nodejs.org)
- [NPM](https://www.npmjs.com/)
- [React](https://reactjs.org/)
- [Faency](https://github.com/containous/faency)
- [React Testing Library](https://testing-library.com/docs/react-testing-library/intro)
