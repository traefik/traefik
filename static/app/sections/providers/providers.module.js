(function () {
  'use strict';

  angular
    .module('traefik.section.providers', [
      'traefik.core.provider',
      'traefik.section.providers.backend-monitor',
      'traefik.section.providers.frontend-monitor'
    ])
    .config(['$stateProvider', function ($stateProvider) {

      $stateProvider.state('provider', {
        url: '/',
        templateUrl: 'app/sections/providers/providers.html',
        controller: 'ProvidersController',
        controllerAs: 'providersCtrl'
      });

    }]);

})();
