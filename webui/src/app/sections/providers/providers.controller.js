'use strict';

/** @ngInject */
function ProvidersController($scope, $interval, $log, Providers) {
  const vm = this;

  function loadProviders() {
    Providers
      .get()
      .then(providers => vm.providers = providers)
      .catch(error => {
        vm.providers = {};
        $log.error(error);
      });
  }

  loadProviders();

  const intervalId = $interval(loadProviders, 2000);

  $scope.$on('$destroy', function () {
    $interval.cancel(intervalId);
  });
}

module.exports = ProvidersController;
