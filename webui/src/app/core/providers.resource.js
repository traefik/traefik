'use strict';
var angular = require('angular');

var traefikCoreProvider = 'traefik.core.provider';
module.exports = traefikCoreProvider;

angular
  .module(traefikCoreProvider, ['ngResource'])
  .factory('Providers', Providers);

/** @ngInject */
function Providers($resource) {
  const resourceProvider = $resource('../api/providers');
  return {
    get: function () {
      const rawProviders = resourceProvider.get();

      for (let providerName in rawProviders) {
        if (rawProviders.hasOwnProperty(providerName)) {

          // BackEnds mapping
          let bckends = rawProviders[providerName].backends;

          rawProviders[providerName].backends = Object.keys(bckends)
            .map(key => {
              const goodBackend = bckends[key];
              goodBackend.backendId = key;
              return goodBackend;
            });

          // FrontEnds mapping
          let frtends = rawProviders[providerName].frontends;

          rawProviders[providerName].frontends = Object.keys(frtends)
            .map(key => {
              const goodFrontend = frtends[key];
              goodFrontend.frontendId = key;
              return goodFrontend;
            });

        }
      }

      return rawProviders;
    }
  };
}
