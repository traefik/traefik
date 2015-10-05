(function () {
  'use strict';

  angular
    .module('traefik.section')
    .config(['$urlRouterProvider', function ($urlRouterProvider) {

      $urlRouterProvider.otherwise('/');

    }]);

})();
