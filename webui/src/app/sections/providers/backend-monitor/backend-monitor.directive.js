(function () {
  'use strict';

  angular
    .module('traefik.section.providers.backend-monitor')
    .directive('backendMonitor', backendMonitor);

    function backendMonitor() {
      return {
        restrict: 'EA',
        templateUrl: 'app/sections/providers/backend-monitor/backend-monitor.html',
        controller: BackendMonitorController,
        controllerAs: 'backendCtrl',
        bindToController: true,
        scope: {
          backend: '=',
          backendId: '='
        }
      };
    }

    function BackendMonitorController() {
      // Nothing
    }

})();
