(function () {
  'use strict';

  angular
    .module('traefik.section', [
      'ui.router',
      'traefik.section.providers',
      'traefik.section.health'
     ]);

})();
