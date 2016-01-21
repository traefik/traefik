(function () {
  'use strict';

  angular
    .module('traefik.core.provider', ['ngResource'])
    .factory('Providers', Providers);

    /** @ngInject */
    function Providers($resource) {
      return $resource('/api/providers');
    }

})();
