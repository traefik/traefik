(function () {
  'use strict';

  angular
    .module('traefik.section', [
      'ui.router',
      'ui.bootstrap',
      'nvd3',
      'traefik.section.providers',
      'traefik.section.health'
     ]);

})();
