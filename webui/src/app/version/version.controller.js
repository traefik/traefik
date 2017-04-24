'use strict';

/** @ngInject */
function VersionController($scope, Version) {
  Version.get(function (version) {
    $scope.version = version;
  });
}

module.exports = VersionController;
