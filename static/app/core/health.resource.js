(function () {
  'use strict';

  angular
    .module('traefik.core.health', ['ngResource'])
    .factory('Health', ['$resource', function ($resource) {
      return $resource('/health');
   }]);

})();
