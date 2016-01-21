(function () {
  'use strict';

  angular
    .module('traefik.section.providers.frontend-monitor')
    .directive('frontendMonitor', frontendMonitor);

    function frontendMonitor() {
      return {
        restrict: 'EA',
        templateUrl: 'app/sections/providers/frontend-monitor/frontend-monitor.html',
        controller: FrontendMonitorController,
        controllerAs: 'frontendCtrl',
        bindToController: true,
        scope: {
          frontend: '=',
          frontendId: '='
        }
      };
    }

    function FrontendMonitorController() {
      // Nothing
    }

})();
