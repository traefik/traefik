(function () {
  'use strict';

    angular
      .module('traefik.core.health', ['ngResource'])
      .factory('Health', Health);

      /** @ngInject */
      function Health($resource) {
        return $resource('/health');
      }

})();
