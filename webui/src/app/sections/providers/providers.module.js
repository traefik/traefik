'use strict';
var angular = require('angular');
var traefikCoreProvider = require('../../core/providers.resource');
var ProvidersController = require('./providers.controller');
var traefikBackendMonitor = require('./backend-monitor/backend-monitor.module');
var traefikFrontendMonitor = require('./frontend-monitor/frontend-monitor.module');

var traefikSectionProviders = 'traefik.section.providers';
module.exports = traefikSectionProviders;

angular
  .module(traefikSectionProviders, [
    traefikCoreProvider,
    traefikBackendMonitor,
    traefikFrontendMonitor
  ])
  .config(config)
  .controller('ProvidersController', ProvidersController);

  /** @ngInject */
  function config($stateProvider) {

    $stateProvider.state('provider', {
      url: '/',
      template: require('./providers.html'),
      controller: 'ProvidersController',
      controllerAs: 'providersCtrl'
    });

  }