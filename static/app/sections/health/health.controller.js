(function () {
  'use strict';

  angular.module('traefik.section.health')
    .controller('HealthController', ['$scope', '$interval', '$log', 'Health', function ($scope, $interval, $log, Health) {

      var vm = this;

      vm.health = Health.get();

      var intervalId = $interval(function () {
        Health.get(function (health) {
          vm.health = health;
        }, function (error) {
          vm.health = {};
          $log.error(error);
        });
      }, 3000);

      $scope.$on('$destroy', function () {
        $interval.cancel(intervalId);
      });

    }]);

})();
