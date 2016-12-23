'use strict';
var angular = require('angular');
var traefikCoreHealth = require('../../core/health.resource');
var HealthController = require('./health.controller');

var traefikSectionHealth = 'traefik.section.health';
module.exports = traefikSectionHealth;

angular
  .module(traefikSectionHealth, [traefikCoreHealth])
  .controller('HealthController', HealthController)
  .config(config);

  /** @ngInject */
  function config($stateProvider) {

    $stateProvider.state('health', {
      url: '/health',
      template: require('./health.html'),
      controller: 'HealthController',
      controllerAs: 'healthCtrl'
    });

  }