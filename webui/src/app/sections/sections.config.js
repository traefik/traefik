(function () {
  'use strict';

  angular
    .module('traefik.section')
    .config(config);

    /** @ngInject */
    function config($urlRouterProvider) {
      $urlRouterProvider.otherwise('/');
    }

})();
