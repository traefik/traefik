'use strict';
var angular = require('angular');

var traefikCoreVersion = 'traefik.core.version';
module.exports = traefikCoreVersion;

angular
  .module(traefikCoreVersion, ['ngResource'])
  .factory('Version', Version);

  /** @ngInject */
  function Version($resource) {
    return $resource('../api/version');
  }
