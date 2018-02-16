'use strict';

var _ = require('lodash');

/** @ngInject */
function ProvidersController($scope, $interval, $log, Providers) {
  const vm = this;

  function loadProviders() {
    Providers
      .get()
      .then(providers => {
        if (!_.isEqual(vm.previousProviders, providers)) {
          vm.providers = providers;
          vm.previousProviders = _.cloneDeep(providers);
        }
      })
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
