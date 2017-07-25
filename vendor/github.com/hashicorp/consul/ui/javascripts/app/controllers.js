App.ApplicationController = Ember.ObjectController.extend({
  updateCurrentPath: function() {
    App.set('currentPath', this.get('currentPath'));
  }.observes('currentPath')
});

App.DcController = Ember.Controller.extend({
  needs: ["application"],
  // Whether or not the dropdown menu can be seen
  isDropdownVisible: false,

  datacenter: Ember.computed.alias('content'),

  // Returns the total number of failing checks.
  //
  // We treat any non-passing checks as failing
  //
  totalChecksFailing: function() {
    return this.get('nodes').reduce(function(sum, node) {
      return sum + node.get('failingChecks');
    }, 0);
  }.property('nodes'),

  totalChecksPassing: function() {
    return this.get('nodes').reduce(function(sum, node) {
      return sum + node.get('passingChecks');
    }, 0);
  }.property('nodes'),

  //
  // Returns the human formatted message for the button state
  //
  checkMessage: function() {
    var failingChecks = this.get('totalChecksFailing');
    var passingChecks = this.get('totalChecksPassing');

    if (this.get('hasFailingChecks') === true) {
      return  failingChecks + ' failing';
    } else {
      return  passingChecks + ' passing';
    }

  }.property('nodes'),

  //
  //
  //
  checkStatus: function() {
    if (this.get('hasFailingChecks') === true) {
      return "failing";
    } else {
      return "passing";
    }

  }.property('nodes'),

  //
  // Boolean if the datacenter has any failing checks.
  //
  hasFailingChecks: Ember.computed.gt('totalChecksFailing', 0),

  actions: {
    // Hide and show the dropdown menu
    toggle: function(item){
      this.toggleProperty('isDropdownVisible');
    },
    // Just hide the dropdown menu
    hideDrop: function(item){
      this.set('isDropdownVisible', false);
    }
  }
});

KvBaseController = Ember.ObjectController.extend({
  getParentKeyRoute: function() {
    if (this.get('isRoot')) {
      return this.get('rootKey');
    }
    return this.get('parentKey');
  },

  transitionToNearestParent: function(parent) {
    var controller = this;
    var rootKey = controller.get('rootKey');
    var dc = controller.get('dc').get('datacenter');
    var token = App.get('settings.token');

    Ember.$.ajax({
      url: (formatUrl(consulHost + '/v1/kv/' + parent + '?keys', dc, token)),
      type: 'GET'
    }).then(function(data) {
      controller.transitionToRoute('kv.show', parent);
    }).fail(function(response) {
      if (response.status === 404) {
        controller.transitionToRoute('kv.show', rootKey);
      }
    });

    controller.set('isLoading', false);
  }
});

App.KvShowController = KvBaseController.extend(Ember.Validations.Mixin, {
  needs: ["dc"],
  dc: Ember.computed.alias("controllers.dc"),
  isLoading: false,

  actions: {
    // Creates the key from the newKey model
    // set on the route.
    createKey: function() {
      this.set('isLoading', true);

      var controller = this;
      var newKey = controller.get('newKey');
      var parentKey = controller.get('parentKey');
      var grandParentKey = controller.get('grandParentKey');
      var dc = controller.get('dc').get('datacenter');
      var token = App.get('settings.token');

      // If we don't have a previous model to base
      // on our parent, or we're not at the root level,
      // add the prefix
      if (parentKey !== undefined && parentKey !== "/") {
        newKey.set('Key', (parentKey + newKey.get('Key')));
      }

      // Put the Key and the Value retrieved from the form
      Ember.$.ajax({
          url: (formatUrl(consulHost + "/v1/kv/" + newKey.get('Key'), dc, token)),
          type: 'PUT',
          data: newKey.get('Value')
      }).then(function(response) {
        // transition to the right place
        if (newKey.get('isFolder') === true) {
          controller.transitionToRoute('kv.show', newKey.get('Key'));
        } else {
          controller.transitionToRoute('kv.edit', newKey.get('Key'));
        }
        controller.set('isLoading', false);
      }).fail(function(response) {
        // Render the error message on the form if the request failed
        controller.set('errorMessage', 'Received error while processing: ' + response.statusText);
      });
    },

    deleteFolder: function() {

      this.set('isLoading', true);
      var controller = this;
      var dc = controller.get('dc').get('datacenter');
      var grandParent = controller.get('grandParentKey');
      var token = App.get('settings.token');

      if (window.confirm("Are you sure you want to delete this folder?")) {
        // Delete the folder
        Ember.$.ajax({
            url: (formatUrl(consulHost + "/v1/kv/" + controller.get('parentKey') + '?recurse', dc, token)),
            type: 'DELETE'
        }).then(function(response) {
          controller.transitionToNearestParent(grandParent);
        }).fail(function(response) {
          // Render the error message on the form if the request failed
          controller.set('errorMessage', 'Received error while processing: ' + response.statusText);
        });
      }
    }
  }
});

App.KvEditController = KvBaseController.extend({
  isLoading: false,
  needs: ["dc"],
  dc: Ember.computed.alias("controllers.dc"),

  actions: {
    // Updates the key set as the model on the route.
    updateKey: function() {
      this.set('isLoading', true);

      var dc = this.get('dc').get('datacenter');
      var key = this.get("model");
      var controller = this;
      var token = App.get('settings.token');

      // Put the key and the decoded (plain text) value
      // from the form.
      Ember.$.ajax({
          url: (formatUrl(consulHost + "/v1/kv/" + key.get('Key'), dc, token)),
          type: 'PUT',
          data: key.get('valueDecoded')
      }).then(function(response) {
        // If success, just reset the loading state.
        controller.set('isLoading', false);
      }).fail(function(response) {
        // Render the error message on the form if the request failed
        controller.set('errorMessage', 'Received error while processing: ' + response.statusText);
      });
    },

    cancelEdit: function() {
      this.set('isLoading', true);
      this.transitionToRoute('kv.show', this.getParentKeyRoute());
      this.set('isLoading', false);
    },

    deleteKey: function() {
      this.set('isLoading', true);

      var controller = this;
      var dc = controller.get('dc').get('datacenter');
      var key = controller.get("model");
      var parent = controller.getParentKeyRoute();
      var token = App.get('settings.token');

      // Delete the key
      Ember.$.ajax({
          url: (formatUrl(consulHost + "/v1/kv/" + key.get('Key'), dc, token)),
          type: 'DELETE'
      }).then(function(data) {
        controller.transitionToNearestParent(parent);
      }).fail(function(response) {
        // Render the error message on the form if the request failed
        controller.set('errorMessage', 'Received error while processing: ' + response.statusText);
      });
    }
  }

});

ItemBaseController = Ember.ArrayController.extend({
  needs: ["dc", "application"],
  queryParams: ["filter", "status", "condensed"],
  dc: Ember.computed.alias("controllers.dc"),
  condensed: true,
  hasExpanded: true,
  filterText: "Filter by name",
  filter: "", // default
  status: "any status", // default
  statuses: ["any status", "passing", "failing"],

  isShowingItem: function() {
    var currentPath = this.get('controllers.application.currentPath');
    return (currentPath === "dc.nodes.show" || currentPath === "dc.services.show");
  }.property('controllers.application.currentPath'),

  filteredContent: function() {
    var filter = this.get('filter');
    var status = this.get('status');

    var items = this.get('items').filter(function(item){
      return item.get('filterKey').toLowerCase().match(filter.toLowerCase());
    });

    switch (status) {
      case "passing":
        return items.filterBy('hasFailingChecks', false);
      case "failing":
        return items.filterBy('hasFailingChecks', true);
      default:
        return items;
    }

  }.property('filter', 'status', 'items.@each'),

  actions: {
    toggleCondensed: function() {
      this.toggleProperty('condensed');
    }
  }
});

App.NodesShowController = Ember.ObjectController.extend({
  needs: ["dc", "nodes"],
  dc: Ember.computed.alias("controllers.dc"),

  actions: {
    invalidateSession: function(sessionId) {
      this.set('isLoading', true);
      var controller = this;
      var node = controller.get('model');
      var dc = controller.get('dc').get('datacenter');
      var token = App.get('settings.token');

      if (window.confirm("Are you sure you want to invalidate this session?")) {
        // Delete the session
        Ember.$.ajax({
            url: (formatUrl(consulHost + "/v1/session/destroy/" + sessionId, dc, token)),
            type: 'PUT'
        }).then(function(response) {
          return Ember.$.getJSON(formatUrl(consulHost + '/v1/session/node/' + node.Node, dc, token)).then(function(data) {
            controller.set('sessions', data);
          });
        }).fail(function(response) {
          // Render the error message on the form if the request failed
          controller.set('errorMessage', 'Received error while processing: ' + response.statusText);
        });
      }
    }
  }
});

App.NodesController = ItemBaseController.extend({
  items: Ember.computed.alias("nodes"),
});

App.ServicesController = ItemBaseController.extend({
  items: Ember.computed.alias("services"),
});

App.AclsController = Ember.ArrayController.extend({
  needs: ["dc", "application"],
  queryParams: ["filter"],
  filterText: "Filter by name or ID",
  searchBar: true,
  newAclButton: true,
  types: ["management", "client"],

  dc: Ember.computed.alias("controllers.dc"),
  items: Ember.computed.alias("acls"),

  filter: "",

  isShowingItem: function() {
    var currentPath = this.get('controllers.application.currentPath');
    return (currentPath === "dc.acls.show");
  }.property('controllers.application.currentPath'),

  filteredContent: function() {
    var filter = this.get('filter');

    var items = this.get('items').filter(function(item, index, enumerable){
      // First try to match on the name
      var nameMatch = item.get('Name').toLowerCase().match(filter.toLowerCase());
      if (nameMatch !== null) {
        return nameMatch;
      } else {
        return item.get('ID').toLowerCase().match(filter.toLowerCase());
      }
    });

    return items;
  }.property('filter', 'items.@each'),

  actions: {
    createAcl: function() {
      this.set('isLoading', true);

      var controller = this;
      var newAcl = controller.get('newAcl');
      var dc = controller.get('dc').get('datacenter');
      var token = App.get('settings.token');

      // Create the ACL
      Ember.$.ajax({
          url: formatUrl(consulHost + '/v1/acl/create', dc, token),
          type: 'PUT',
          data: JSON.stringify(newAcl)
      }).then(function(response) {
        // transition to the acl
        controller.transitionToRoute('acls.show', response.ID);

        // Get the ACL again, including the newly created one
        Ember.$.getJSON(formatUrl(consulHost + '/v1/acl/list', dc, token)).then(function(data) {
          var objs = [];
          data.map(function(obj){
            objs.push(App.Acl.create(obj));
          });
          controller.set('items', objs);
        });

        controller.set('isLoading', false);
      }).fail(function(response) {
        // Render the error message on the form if the request failed
        notify('Received error while creating ACL: ' + response.statusText, 8000);
        controller.set('isLoading', false);
      });
    },
  }
});


App.AclsShowController = Ember.ObjectController.extend({
  needs: ["dc", "acls"],
  dc: Ember.computed.alias("controllers.dc"),
  isLoading: false,
  types: ["management", "client"],

  actions: {
    set: function() {
      this.set('isLoading', true);
      var controller = this;
      var acl = controller.get('model');
      var dc = controller.get('dc').get('datacenter');

      if (window.confirm("Are you sure you want to use this token for your session?")) {
        // Set
        var token = App.set('settings.token', acl.ID);
        controller.transitionToRoute('services');
        this.set('isLoading', false);
        notify('Now using token: ' + acl.ID, 3000);
      }
    },

    clone: function() {
      this.set('isLoading', true);
      var controller = this;
      var acl = controller.get('model');
      var dc = controller.get('dc').get('datacenter');
      var token = App.get('settings.token');

      // Set
      controller.transitionToRoute('services');

      Ember.$.ajax({
          url: formatUrl(consulHost + '/v1/acl/clone/'+ acl.ID, dc, token),
          type: 'PUT'
      }).then(function(response) {
        controller.transitionToRoute('acls.show', response.ID);
        controller.set('isLoading', false);
        notify('Successfully cloned token', 4000);
      }).fail(function(response) {
        // Render the error message on the form if the request failed
        controller.set('errorMessage', 'Received error while processing: ' + response.statusText);
        controller.set('isLoading', false);
      });

    },

    delete: function() {
      this.set('isLoading', true);
      var controller = this;
      var acl = controller.get('model');
      var dc = controller.get('dc').get('datacenter');
      var token = App.get('settings.token');

      if (window.confirm("Are you sure you want to delete this token?")) {
        Ember.$.ajax({
            url: formatUrl(consulHost + '/v1/acl/destroy/'+ acl.ID, dc, token),
            type: 'PUT'
        }).then(function(response) {
          Ember.$.getJSON(formatUrl(consulHost + '/v1/acl/list', dc, token)).then(function(data) {
            objs = [];
            data.map(function(obj){
              if (obj.ID === "anonymous") {
                objs.unshift(App.Acl.create(obj));
              } else {
                objs.push(App.Acl.create(obj));
              }
            });
            controller.get('controllers.acls').set('acls', objs);
          }).then(function() {
            controller.transitionToRoute('acls');
            controller.set('isLoading', false);
            notify('ACL deleted successfully', 3000);
          });
        }).fail(function(response) {
          // Render the error message on the form if the request failed
          controller.set('errorMessage', 'Received error while processing: ' + response.statusText);
          controller.set('isLoading', false);
        });
      }
    },

    updateAcl: function() {
      this.set('isLoading', true);

      var controller = this;
      var acl = controller.get('model');
      var dc = controller.get('dc').get('datacenter');
      var token = App.get('settings.token');

      // Update the ACL
      Ember.$.ajax({
          url: formatUrl(consulHost + '/v1/acl/update', dc, token),
          type: 'PUT',
          data: JSON.stringify(acl)
      }).then(function(response) {
        // transition to the acl
        controller.set('isLoading', false);
        notify('ACL updated successfully', 3000);
      }).fail(function(response) {
        // Render the error message on the form if the request failed
        notify('Received error while updating ACL: ' + response.statusText, 8000);
        controller.set('isLoading', false);
      });
    }
  }
});

App.SettingsController = Ember.ObjectController.extend({
  actions: {
    reset: function() {
      this.set('isLoading', true);
      var controller = this;

      if (window.confirm("Are your sure you want to reset your settings?")) {
        localStorage.clear();
        controller.set('content', App.Settings.create());
        App.set('settings.token', '');
        notify('Settings reset', 3000);
        this.set('isLoading', false);
      }
    }
  }
});

