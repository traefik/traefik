(function() {
  'use strict';

  angular
    .module('traefik')
    .run(runBlock);

  /** @ngInject */
  function runBlock($log) {

    $log.debug('runBlock end');
  }

})();
