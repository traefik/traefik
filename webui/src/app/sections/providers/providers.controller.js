(function () {
  'use strict';

  angular
    .module('traefik.section.providers')
    .controller('ProvidersController', ProvidersController);

    /** @ngInject */
    function ProvidersController($scope, $interval, $log, Providers) {
      var vm = this;

      vm.providers = Providers.get();

      var intervalId = $interval(function () {
        Providers.get(function (providers) {
          vm.providers = providers;
        }, function (error) {
          vm.providers = {};
          $log.error(error);
        });
      }, 2000);

      $scope.$on('$destroy', function () {
        $interval.cancel(intervalId);
      });
    }

})();
