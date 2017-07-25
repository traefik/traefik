//
// Superclass to be used by all of the main routes below.
//
App.BaseRoute = Ember.Route.extend({
  rootKey: '',
  condensedView: false,

  // Don't record characters in browser history
  // for the "search" query item (filter)
  queryParams: {
    filter: {
      replace: true
    }
  },

  getParentAndGrandparent: function(key) {
    var parentKey = this.rootKey,
        grandParentKey = this.rootKey,
        parts = key.split('/');

    if (parts.length > 0) {
      parts.pop();
      parentKey = parts.join("/") + "/";
    }

    if (parts.length > 1) {
      parts.pop();
      grandParentKey = parts.join("/") + "/";
    }

    return {
      parent: parentKey,
      grandParent: grandParentKey,
      isRoot: parentKey === '/'
    };
  },

  removeDuplicateKeys: function(keys, matcher) {
    // Loop over the keys
    keys.forEach(function(item, index) {
      if (item.get('Key') == matcher) {
      // If we are in a nested folder and the folder
      // name matches our position, remove it
        keys.splice(index, 1);
      }
    });
    return keys;
  },

  actions: {
    // Used to link to keys that are not objects,
    // like parents and grandParents
    linkToKey: function(key) {
      if (key == "/") {
        this.transitionTo('kv.show', "");
      }
      else if (key.slice(-1) === '/' || key === this.rootKey) {
        this.transitionTo('kv.show', key);
      } else {
        this.transitionTo('kv.edit', key);
      }
    }
  }
});

//
// The route for choosing datacenters, typically the first route loaded.
//
App.IndexRoute = App.BaseRoute.extend({
  // Retrieve the list of datacenters
  model: function(params) {
    return Ember.$.getJSON(consulHost + '/v1/catalog/datacenters').then(function(data) {
      return data;
    });
  },

  afterModel: function(model, transition) {
    // If we only have one datacenter, jump
    // straight to it and bypass the global
    // view
    if (model.get('length') === 1) {
      this.transitionTo('services', model[0]);
    }
  }
});

// The parent route for all resources. This keeps the top bar
// functioning, as well as the per-dc requests.
App.DcRoute = App.BaseRoute.extend({
  model: function(params) {
    var token = App.get('settings.token');

    // Return a promise hash to retreieve the
    // dcs and nodes used in the header
    return Ember.RSVP.hash({
      dc: params.dc,
      dcs: Ember.$.getJSON(consulHost + '/v1/catalog/datacenters'),
      nodes: Ember.$.getJSON(formatUrl(consulHost + '/v1/internal/ui/nodes', params.dc, token)).then(function(data) {
        var objs = [];

        // Merge the nodes into a list and create objects out of them
        data.map(function(obj){
          objs.push(App.Node.create(obj));
        });

        return objs;
      }),
      coordinates: Ember.$.getJSON(formatUrl(consulHost + '/v1/coordinate/nodes', params.dc, token)).then(function(data) {
        return data;
      })
    });
  },

  setupController: function(controller, models) {
    controller.set('content', models.dc);
    controller.set('nodes', models.nodes);
    controller.set('dcs', models.dcs);
    controller.set('coordinates', models.coordinates);
    controller.set('isDropdownVisible', false);
  },
});

App.KvIndexRoute = App.BaseRoute.extend({
  beforeModel: function() {
    this.transitionTo('kv.show', this.rootKey);
  }
});

App.KvShowRoute = App.BaseRoute.extend({
  model: function(params) {
    var key = params.key;
    var dc = this.modelFor('dc').dc;
    var token = App.get('settings.token');

    // Return a promise has with the ?keys for that namespace
    // and the original key requested in params
    return Ember.RSVP.hash({
      key: key,
      keys: Ember.$.getJSON(formatUrl(consulHost + '/v1/kv/' + key + '?keys&seperator=/', dc, token)).then(function(data) {
        var objs = [];
        data.map(function(obj){
          objs.push(App.Key.create({Key: obj}));
        });
        return objs;
      })
    });
  },

  setupController: function(controller, models) {
    var key = models.key;
    var parentKeys = this.getParentAndGrandparent(key);
    models.keys = this.removeDuplicateKeys(models.keys, models.key);

    controller.set('content', models.keys);
    controller.set('parentKey', parentKeys.parent);
    controller.set('grandParentKey', parentKeys.grandParent);
    controller.set('isRoot', parentKeys.isRoot);
    controller.set('newKey', App.Key.create());
    controller.set('rootKey', this.rootKey);
  }
});

App.KvEditRoute = App.BaseRoute.extend({
  model: function(params) {
    var key = params.key;
    var dc = this.modelFor('dc').dc;
    var parentKeys = this.getParentAndGrandparent(key);
    var token = App.get('settings.token');

    // Return a promise hash to get the data for both columns
    return Ember.RSVP.hash({
      dc: dc,
      token: token,
      key: Ember.$.getJSON(formatUrl(consulHost + '/v1/kv/' + key, dc, token)).then(function(data) {
        // Convert the returned data to a Key
        return App.Key.create().setProperties(data[0]);
      }),
      keys: keysPromise = Ember.$.getJSON(formatUrl(consulHost + '/v1/kv/' + parentKeys.parent + '?keys&seperator=/', dc, token)).then(function(data) {
        var objs = [];
        data.map(function(obj){
         objs.push(App.Key.create({Key: obj}));
        });
        return objs;
      }),
    });
  },

  // Load the session on the key, if there is one
  afterModel: function(models) {
    if (models.key.get('isLocked')) {
      return Ember.$.getJSON(formatUrl(consulHost + '/v1/session/info/' + models.key.Session, models.dc, models.token)).then(function(data) {
        models.session = data[0];
        return models;
      });
    } else {
      return models;
    }
  },

  setupController: function(controller, models) {
    var key = models.key;
    var parentKeys = this.getParentAndGrandparent(key.get('Key'));
    models.keys = this.removeDuplicateKeys(models.keys, parentKeys.parent);

    controller.set('content', models.key);
    controller.set('parentKey', parentKeys.parent);
    controller.set('grandParentKey', parentKeys.grandParent);
    controller.set('isRoot', parentKeys.isRoot);
    controller.set('siblings', models.keys);
    controller.set('rootKey', this.rootKey);
    controller.set('session', models.session);
  }
});

App.ServicesRoute = App.BaseRoute.extend({
  model: function(params) {
    var dc = this.modelFor('dc').dc;
    var token = App.get('settings.token');

    // Return a promise to retrieve all of the services
    return Ember.$.getJSON(formatUrl(consulHost + '/v1/internal/ui/services', dc, token)).then(function(data) {
      var objs = [];
      data.map(function(obj){
       objs.push(App.Service.create(obj));
      });
      return objs;
    });
  },
  setupController: function(controller, model) {
    controller.set('services', model);
  }
});


App.ServicesShowRoute = App.BaseRoute.extend({
  model: function(params) {
    var dc = this.modelFor('dc').dc;
    var token = App.get('settings.token');

    // Here we just use the built-in health endpoint, as it gives us everything
    // we need.
    return Ember.$.getJSON(formatUrl(consulHost + '/v1/health/service/' + params.name, dc, token)).then(function(data) {
      var objs = [];
      data.map(function(obj){
       objs.push(App.Node.create(obj));
      });
      return objs;
    });
  },
  setupController: function(controller, model) {
    var tags = [];
    model.map(function(obj){
      if (obj.Service.Tags !== null) {
        tags = tags.concat(obj.Service.Tags);
      }
    });

    tags = tags.filter(function(n){ return n !== undefined; });
    tags = tags.uniq().join(', ');

    controller.set('content', model);
    controller.set('tags', tags);
  }
});

function distance(a, b) {
    a = a.Coord;
    b = b.Coord;
    var sum = 0;
    for (var i = 0; i < a.Vec.length; i++) {
        var diff = a.Vec[i] - b.Vec[i];
        sum += diff * diff;
    }
    var rtt = Math.sqrt(sum) + a.Height + b.Height;

    var adjusted = rtt + a.Adjustment + b.Adjustment;
    if (adjusted > 0.0) {
        rtt = adjusted;
    }

    return Math.round(rtt * 100000.0) / 100.0;
}

App.NodesShowRoute = App.BaseRoute.extend({
  model: function(params) {
    var dc = this.modelFor('dc');
    var token = App.get('settings.token');

    var min = 999999999;
    var max = -999999999;
    var sum = 0;
    var distances = [];
    dc.coordinates.forEach(function (node) {
      if (params.name == node.Node) {
        dc.coordinates.forEach(function (other) {
          if (node.Node != other.Node) {
            var dist = distance(node, other);
            distances.push({ node: other.Node, distance: dist });
            sum += dist;
            if (dist < min) {
              min = dist;
            }
            if (dist > max) {
              max = dist;
            }
          }
        });
        distances.sort(function (a, b) {
          return a.distance - b.distance;
        });
      }
    });
    var n = distances.length;
    var halfN = Math.floor(n / 2);
    var median;

    if (n > 0) {
      if (n % 2) {
        // odd
        median = distances[halfN].distance;
      } else {
        median = (distances[halfN - 1].distance + distances[halfN].distance) / 2;
      }
    } else {
      median = 0;
      min = 0;
      max = 0;
    }

    // Return a promise hash of the node and nodes
    return Ember.RSVP.hash({
      dc: dc.dc,
      token: token,
      tomography: {
        distances: distances,
        n: distances.length,
        min: parseInt(min * 100) / 100,
        median: parseInt(median * 100) / 100,
        max: parseInt(max * 100) / 100
      },
      node: Ember.$.getJSON(formatUrl(consulHost + '/v1/internal/ui/node/' + params.name, dc.dc, token)).then(function(data) {
        return App.Node.create(data);
      }),
      nodes: Ember.$.getJSON(formatUrl(consulHost + '/v1/internal/ui/node/' + params.name, dc.dc, token)).then(function(data) {
        return App.Node.create(data);
      })
    });
  },

  // Load the sessions for the node
  afterModel: function(models) {
    return Ember.$.getJSON(formatUrl(consulHost + '/v1/session/node/' + models.node.Node, models.dc, models.token)).then(function(data) {
      models.sessions = data;
      return models;
    });
  },

  setupController: function(controller, models) {
      controller.set('content', models.node);
      controller.set('sessions', models.sessions);
      controller.set('tomography', models.tomography);
      //
      // Since we have 2 column layout, we need to also display the
      // list of nodes on the left. Hence setting the attribute
      // {{nodes}} on the controller.
      //
      controller.set('nodes', models.nodes);
  }
});

App.NodesRoute = App.BaseRoute.extend({
  model: function(params) {
    var dc = this.modelFor('dc').dc;
    var token = App.get('settings.token');

    // Return a promise containing the nodes
    return Ember.$.getJSON(formatUrl(consulHost + '/v1/internal/ui/nodes', dc, token)).then(function(data) {
      var objs = [];
      data.map(function(obj){
       objs.push(App.Node.create(obj));
      });
      return objs;
    });
  },
  setupController: function(controller, model) {
    controller.set('nodes', model);
  }
});


App.AclsRoute = App.BaseRoute.extend({
  model: function(params) {
    var dc = this.modelFor('dc').dc;
    var token = App.get('settings.token');
    // Return a promise containing the ACLS
    return Ember.$.getJSON(formatUrl(consulHost + '/v1/acl/list', dc, token)).then(function(data) {
      var objs = [];
      data.map(function(obj){
        if (obj.ID === "anonymous") {
          objs.unshift(App.Acl.create(obj));
        } else {
          objs.push(App.Acl.create(obj));
        }
      });
      return objs;
    });
  },

  actions: {
    error: function(error, transition) {
      // If consul returns 401, ACLs are disabled
      if (error && error.status === 401) {
        this.transitionTo('dc.aclsdisabled');
      // If consul returns 403, they key isn't authorized for that
      // action.
      } else if (error && error.status === 403) {
        this.transitionTo('dc.unauthorized');
      }
      return true;
    }
  },

  setupController: function(controller, model) {
      controller.set('acls', model);
      controller.set('newAcl', App.Acl.create());
  }
});

App.AclsShowRoute = App.BaseRoute.extend({
  model: function(params) {
    var dc = this.modelFor('dc').dc;
    var token = App.get('settings.token');

    // Return a promise hash of the node and nodes
    return Ember.RSVP.hash({
      dc: dc,
      acl: Ember.$.getJSON(formatUrl(consulHost + '/v1/acl/info/'+ params.id, dc, token)).then(function(data) {
        return App.Acl.create(data[0]);
      })
    });
  },

  setupController: function(controller, models) {
      controller.set('content', models.acl);
  }
});

App.SettingsRoute = App.BaseRoute.extend({
  model: function(params) {
    return App.get('settings');
  }
});


// Adds any global parameters we need to set to a url/path
function formatUrl(url, dc, token) {
  if (token == null) {
    token = "";
  }
  if (url.indexOf("?") > 0) {
    // If our url has existing params
    url = url + "&dc=" + dc;
    url = url + "&token=" + token;
  } else {
    // Our url doesn't have params
    url = url + "?dc=" + dc;
    url = url + "&token=" + token;
  }
  return url;
}
