'use strict';
var d3 = require('d3');

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
      transitionDuration: 50,
      yAxis: {
        axisLabelDistance: 30
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
      transitionDuration: 50,
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
      }
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
   * Load all graph's datas
   *
   * @param {Object} health Health data from server
   */
  function loadData(health) {
    // Load datas and update Average Response Time graph render
    updateAverageResponseTimeGraph(health.unixtime, health.average_response_time_sec);

    // Load datas and update Total Status Code Count graph render
    updateTotalStatusCodeCount(health.total_status_code_count);

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
