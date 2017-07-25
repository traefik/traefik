// in order to see the app running inside the QUnit runner
App.rootElement = '#ember-testing';

// Common test setup
App.setupForTesting();
App.injectTestHelpers();

// Test "fixtures". We populate these based on the running consul
// on the machine where you run the tests.
var fixtures = {
  dc: "dc1",
  node: null,
  service: null,
  key: "fake",
  value: "foobar"
}

module("Integration tests", {
  setup: function() {
    // before each test, ensure the application is ready to run.
    Ember.run(App, App.advanceReadiness);

    // Discover the service, node and dc info
    Ember.$.getJSON('/v1/catalog/datacenters').then(function(data) {
      fixtures.dc = data[0]
    }).then(function(){
      Ember.$.getJSON('/v1/internal/ui/nodes?dc=' + fixtures.dc).then(function(data) {
        fixtures.node = data[0].Node
      });
    }).then(function(){
      Ember.$.getJSON('/v1/internal/ui/services?dc=' + fixtures.dc).then(function(data) {
        fixtures.service = data[0].Name
      });
    });
    // Create a fake key
    Ember.$.ajax({
      url: ("/v1/kv/" + fixtures.key + '?dc=' + fixtures.dc),
      type: 'PUT',
      data: fixtures.value
    })
  },

  teardown: function() {
    // reset the application state between each test
    App.reset();
  }
});

test("services", function() {
  visit("/")

  andThen(function() {
    ok(find("a:contains('Services')").hasClass('active'), "highlights services in nav");
    equal(find(".ember-list-item-view").length, 1, "renders one service");
    ok(find(".ember-list-item-view .name:contains('"+ fixtures.service +"')"), "uses service name");
    ok(find(".ember-list-item-view .name:contains('passing')"), "shows passing check num");
  });
});

test("servicesShow", function() {
  visit("/");
  // First item in list
  click('.ember-list-item-view .list-group-item');

  andThen(function() {
    ok(find("a:contains('Services')").hasClass('active'), "highlights services in nav");
    equal(find(".ember-list-item-view").length, 1, "renders one service");
    ok(find(".ember-list-item-view .list-group-item").hasClass('active'), "highlights active service");
    ok(find(".ember-list-item-view .name:contains('"+ fixtures.service +"')"), "uses service name");
    ok(find(".ember-list-item-view .name:contains('passing')"), "shows passing check num");
    ok(find("h3:contains('"+ fixtures.service+"')"), "shows service name");
    equal(find("h5").text(), "Nodes", "shows node list");
    ok(find("h3.panel-title:contains('"+ fixtures.node +"')"), "shows node name");
  });
});

test("nodes", function() {
  visit("/");
  click("a:contains('Nodes')");

  andThen(function() {
    ok(find("a:contains('Nodes')").hasClass('active'), "highlights nodes in nav");
    equal(find(".ember-list-item-view").length, 1, "renders one node");
    ok(find(".ember-list-item-view .name:contains('"+ fixtures.node +"')"), "contains node name");
    ok(find(".ember-list-item-view .name:contains('services')"), "contains services num");
  });
});

test("nodesShow", function() {
  visit("/");
  click("a:contains('Nodes')");
  // First item in list
  click('.ember-list-item-view .list-group-item');

  andThen(function() {
    ok(find("a:contains('Nodes')").hasClass('active'), "highlights services in nav");
    equal(find(".ember-list-item-view").length, 1, "renders one service");
    ok(find(".ember-list-item-view .list-group-item").hasClass('active'), "highlights active node");
    ok(find(".ember-list-item-view .name:contains('"+ fixtures.node +"')"), "uses node name");
    ok(find(".ember-list-item-view .name:contains('passing')"), "shows passing check num");
  });
});

test("kv", function() {
  visit("/");
  click("a:contains('Key/Value')");

  andThen(function() {
    ok(find("a:contains('Key/Value')").hasClass('active'), "highlights kv in nav");
    equal(find(".list-group-item").length, 1, "renders one key");
    ok(find(".list-group-item:contains('"+ fixtures.key +"')"), "contains key name");
  });
});
