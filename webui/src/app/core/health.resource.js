'use strict';
var angular = require('angular');

var traefikCoreHealth = 'traefik.core.health';
module.exports = traefikCoreHealth;

angular
  .module(traefikCoreHealth, ['ngResource'])
  .factory('Health', Health);

  /** @ngInject */
  function Health($resource) {
    return $resource('../health');
  }