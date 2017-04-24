'use strict';

function backendMonitor() {
  return {
    restrict: 'EA',
    template: require('./backend-monitor.html'),
    controller: BackendMonitorController,
    controllerAs: 'backendCtrl',
    bindToController: true,
    scope: {
      backend: '='
    }
  };
}

function BackendMonitorController() {
  // Nothing
}

module.exports = backendMonitor;
