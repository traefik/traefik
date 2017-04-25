'use strict';

/** @ngInject */
function ProvidersController($scope, $interval, $log, Providers) {
  const vm = this;

  vm.providers = Providers.get();

  const intervalId = $interval(function () {
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

module.exports = ProvidersController;
