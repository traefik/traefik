'use strict';
var angular = require('angular');
var ngAnimate = require('angular-animate');
var ngCookies = require('angular-cookies');
var ngSanitize = require('angular-sanitize');
var ngMessages = require('angular-messages');
var ngAria = require('angular-aria');
var ngResource = require('angular-resource');
var uiRouter = require('angular-ui-router');
var uiBootstrap = require('angular-ui-bootstrap');
var moment = require('moment');
var traefikSection = require('./app/sections/sections');
var traefikVersion = require('./app/version/version.module');
require('./index.scss');
require('animate.css/animate.css');
require('nvd3/build/nv.d3.css');
require('bootstrap/dist/css/bootstrap.css');

var app = 'traefik';
module.exports = app;

angular
  .module(app, [
    ngAnimate,
    ngCookies,
    ngSanitize,
    ngMessages,
    ngAria,
    ngResource,
    uiRouter,
    uiBootstrap,
    traefikSection,
    traefikVersion
  ])
  .run(runBlock)
  .constant('moment', moment)
  .config(config);

/** @ngInject */
function config($logProvider) {
  // Enable log
  $logProvider.debugEnabled(true);
}

/** @ngInject */
function runBlock($log) {
  $log.debug('runBlock end');
}
