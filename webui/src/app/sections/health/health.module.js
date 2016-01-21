(function () {
  'use strict';

  angular.module('traefik.section.health', ['traefik.core.health'])
    .config(config);

    /** @ngInject */
    function config($stateProvider) {

      $stateProvider.state('health', {
        url: '/health',
        templateUrl: 'app/sections/health/health.html',
        controller: 'HealthController',
        controllerAs: 'healthCtrl'
      });

    }

})();
