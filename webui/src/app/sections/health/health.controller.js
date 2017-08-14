'use strict';
var d3 = require('d3'),
    moment = require('moment');

/** @ngInject */
function HealthController($scope, $interval, $log, Health) {

  var vm = this;

  vm.graph = {
    averageResponseTime: {},
    totalStatusCodeCount: {}
  };

  vm.graph.totalStatusCodeCount.options = {
    "chart": {
      type: 'discreteBarChart',
      height: 200,
      margin: {
        top: 20,
        right: 20,
        bottom: 40,
        left: 55
      },
      x: function (d) {
        return d.label;
      },
      y: function (d) {
        return d.value;
      },
      showValues: true,
      valueFormat: function (d) {
        return d3.format('d')(d);
      },
      yAxis: {
        axisLabelDistance: 30,
        tickFormat: d3.format('d')
      }
    },
    "title": {
      "enable": true,
      "text": "Total Status Code Count",
      "css": {
        "textAlign": "center"
      }
    }
  };

  vm.graph.totalStatusCodeCount.data = [
    {
      key: "Total Status Code Count",
      values: [
        {
          "label": "200",
          "value": 0
        }
      ]
    }
  ];

  /**
   * Update Total Status Code Count graph
   *
   * @param {Object} totalStatusCodeCount Object from API
   */
  function updateTotalStatusCodeCount(totalStatusCodeCount) {

    // extract values
    vm.graph.totalStatusCodeCount.data[0].values = [];
    for (var code in totalStatusCodeCount) {
      if (totalStatusCodeCount.hasOwnProperty(code)) {
        vm.graph.totalStatusCodeCount.data[0].values.push({
          label: code,
          value: totalStatusCodeCount[code]
        });
      }
    }

    // Update Total Status Code Count graph render
    if (vm.graph.totalStatusCodeCount.api) {
      vm.graph.totalStatusCodeCount.api.update();
    } else {
      $log.error('fail');
    }

  }

  vm.graph.averageResponseTime.options = {
    chart: {
      type: 'lineChart',
      height: 200,
      margin: {
        top: 20,
        right: 40,
        bottom: 40,
        left: 55
      },
      x: function (d) {
        return d.x;
      },
      y: function (d) {
        return d.y;
      },
      useInteractiveGuideline: true,
      xAxis: {
        tickFormat: function (d) {
          return d3.time.format('%X')(new Date(d));
        }
      },
      yAxis: {
        tickFormat: function (d) {
          return d3.format(',.1f')(d);
        }
      },
      forceY: [0., 1.], // This prevents the chart from showing -1 on Oy when all the input data points
                        // have y = 0. It won't disable the automatic adjustment of the max value.
      duration: 0 // Bug: Markers will not be drawn if you set this to some other value...
    },
    "title": {
      "enable": true,
      "text": "Average response time",
      "css": {
        "textAlign": "center"
      }
    }
  };

  var initialPoint = {
    x: Date.now() - 3000,
    y: 0
  };
  vm.graph.averageResponseTime.data = [
    {
      values: [initialPoint],
      key: 'Average response time (ms)',
      type: 'line',
      color: '#2ca02c'
    }
  ];

  /**
   * Update average response time graph
   *
   * @param {Number} x     Coordinate X
   * @param {Number} y     Coordinate Y
   */
  function updateAverageResponseTimeGraph(x, y) {

    // x multiply 1000 by because unix time is in seconds and JS Date are in milliseconds
    var data = {
      x: x * 1000,
      y: y * 1000
    };
    vm.graph.averageResponseTime.data[0].values.push(data);
    // limit graph entries
    if (vm.graph.averageResponseTime.data[0].values.length > 100) {
      vm.graph.averageResponseTime.data[0].values.shift();
    }

    // Update Average Response Time graph render
    if (vm.graph.averageResponseTime.api) {
      vm.graph.averageResponseTime.api.update();
    }
  }

  /**
   * Format the timestamp as "x seconds ago", etc.
   *
   * @param {String} t Timestamp returned from the API
   */
  function formatTimestamp(t) {
    return moment(t, "YYYY-MM-DDTHH:mm:ssZ").fromNow();
  }

  /**
   * Load all graph's datas
   *
   * @param {Object} health Health data from server
   */
  function loadData(health) {
    // Load datas and update Average Response Time graph render
    updateAverageResponseTimeGraph(health.unixtime, health.average_response_time_sec);

    // Load datas and update Total Status Code Count graph render
    updateTotalStatusCodeCount(health.total_status_code_count);

    // Format the timestamps
    if (health.recent_errors) {
      angular.forEach(health.recent_errors, function(i) {
        i.time_formatted = formatTimestamp(i.time);
      });
    }

    // set data's view
    vm.health = health;
  }

  /**
   * Action when load datas failed
   *
   * @param {Object} error Error state object
   */
  function erroData(error) {
    vm.health = {};
    $log.error(error);
  }

  // first load
  Health.get(loadData, erroData);

  // Auto refresh data
  var intervalId = $interval(function () {
    Health.get(loadData, erroData);
  }, 3000);

  // Stop auto refresh when page change
  $scope.$on('$destroy', function () {
    $interval.cancel(intervalId);
  });

}

module.exports = HealthController;
