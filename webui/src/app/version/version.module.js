'use strict';
var angular = require('angular');
var traefikCoreVersion = require('../core/version.resource');
var VersionController = require('./version.controller');

var traefikVersion = 'traefik.version';
module.exports = traefikVersion;

angular
  .module(traefikVersion, [traefikCoreVersion])
  .controller('VersionController', VersionController);
