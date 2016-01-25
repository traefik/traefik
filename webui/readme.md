# Træfɪk Web UI

Access to Træfɪk Web UI, ex: http://localhost:8080

## Interface

Træfɪk Web UI provide 2 types of informations:
- Providers with their backends and frontends information.
- Health of the web server.

## How to build (for backends developer)

Use the make file :

```shell
make build           # Generate Docker image
make generate-webui  # Generate static contents in `traefik/static/` folder.
```

## How to build (only for frontends developer)

- prerequisite: [Node](https://nodejs.org)

- Go to the directory `webui`

- To install dependencies, execute the following commands:
  - `npm install`
  - `bower install`

- Build static Web UI, execute the following command:
  - `gulp`

- Static contents are build in the directory `static`

**Don't change manually the files in the directory `static`**

- The build allow to:
  - optimize all JavaScript
  - optimize all CSS
  - add vendor prefixes to CSS (cross-bowser support)
  - add a hash in the file names to prevent browser cache problems
  - all images will be optimized at build


## How to edit (only for frontends developer)

**Don't change manually the files in the directory `static`**

- Go to the directory `webui`
- Edit files in `webui/src`

- Run in development mode :
  - `gulp serve`

- Træfɪk API connections are defined in:
  - `webui/src/app/core/health.resource.js`
  - `webui/src/app/core/providers.resource.js`

- The pages contents are in the directory `webui/src/app/sections`.


## Libraries

- [Node](https://nodejs.org)
- [Generator Gulp-Angular](https://github.com/Swiip/generator-gulp-angular)
- [AngularJS](https://docs.angularjs.org/api)
- [UI Router](https://github.com/angular-ui/ui-router)
  - [UI Router - Documentation](https://github.com/angular-ui/ui-router/wiki)
- [Bootstrap](http://getbootstrap.com)
- [Angular Bootstrap](https://angular-ui.github.io/bootstrap)
- [D3](http://d3js.org)
  - [D3 - Documentation](https://github.com/mbostock/d3/wiki)
- [NVD3](http://nvd3.org)
- [Angular nvD3](http://krispo.github.io/angular-nvd3)
