window.App = Ember.Application.create({
  rootElement: "#app",
  currentPath: ''
});

Ember.Application.initializer({
  name: 'settings',

  initialize: function(container, application) {
    application.set('settings', App.Settings.create());
    if (App.get('settings.token') === undefined) {
      App.set('settings.token', '');
    }
  }
});

App.Router.map(function() {
  // Our parent datacenter resource sets the namespace
  // for the entire application
  this.resource("dc", {path: "/:dc"}, function() {
    // Services represent a consul service
    this.resource("services", { path: "/services" }, function(){
      // Show an individual service
      this.route("show", { path: "/*name" });
    });
    // Nodes represent a consul node
    this.resource("nodes", { path: "/nodes" }, function() {
      // Show an individual node
      this.route("show", { path: "/:name" });
    });
    // Key/Value
    this.resource("kv", { path: "/kv" }, function(){
      this.route("index", { path: "/" });
      // List keys. This is more like an index
      this.route("show", { path: "/*key" });
      // Edit a specific key
      this.route("edit", { path: "/*key/edit" });
    });
    // ACLs
    this.resource("acls", { path: "/acls" }, function(){
      this.route("show", { path: "/:id" });
    });

    // Shows a page explaining that ACLs haven't been set-up
    this.route("aclsdisabled", { path: "/aclsdisabled" });
    // Shows a page explaining that the ACL key being used isn't
    // authorized
    this.route("unauthorized", { path: "/unauthorized" });

    this.resource("settings", { path: "/settings" });
  });

  // Shows a datacenter picker. If you only have one
  // it just redirects you through.
  this.route("index", { path: "/" });
});

