'use strict';
var angular = require('angular');
var backendMonitor = require('./backend-monitor.directive');

var traefikBackendMonitor = 'traefik.section.providers.backend-monitor';
module.exports = traefikBackendMonitor;

angular
  .module(traefikBackendMonitor, [])
  .directive('backendMonitor', backendMonitor);
