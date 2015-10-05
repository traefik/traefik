(function () {
  'use strict';

  angular
    .module('traefik.core.provider', ['ngResource'])
    .factory('Providers', ['$resource', function ($resource) {
      return $resource('/api/providers');
   }]);

})();
