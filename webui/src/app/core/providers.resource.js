'use strict';
var angular = require('angular');

var traefikCoreProvider = 'traefik.core.provider';
module.exports = traefikCoreProvider;

angular
  .module(traefikCoreProvider, ['ngResource'])
  .factory('Providers', Providers);

  /** @ngInject */
  function Providers($resource) {
    return $resource('../api/providers');
  }