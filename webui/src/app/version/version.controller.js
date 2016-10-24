'use strict';

/** @ngInject */
function VersionController($scope, $interval, $log, Version) {
  Version.get(function (version) {
    $scope.version = version;
  });
}

module.exports = VersionController;
