'use strict';
var angular = require('angular');
var frontendMonitor = require('./frontend-monitor.directive');

var traefikFrontendMonitor = 'traefik.section.providers.frontend-monitor';
module.exports = traefikFrontendMonitor;

angular
  .module(traefikFrontendMonitor, [])
  .directive('frontendMonitor', frontendMonitor);
