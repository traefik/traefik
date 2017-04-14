'use strict';

function frontendMonitor() {
  return {
    restrict: 'EA',
    template: require('./frontend-monitor.html'),
    controller: FrontendMonitorController,
    controllerAs: 'frontendCtrl',
    bindToController: true,
    scope: {
      frontend: '='
    }
  };
}

function FrontendMonitorController() {
  // Nothing
}

module.exports = frontendMonitor;
