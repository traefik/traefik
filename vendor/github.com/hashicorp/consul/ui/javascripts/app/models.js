//
// A Consul service.
//
App.Service = Ember.Object.extend({
  //
  // The number of failing checks within the service.
  //
  failingChecks: function() {
    // If the service was returned from `/v1/internal/ui/services`
    // then we have a aggregated value which we can just grab
    if (this.get('ChecksCritical') !== undefined) {
      return (this.get('ChecksCritical') + this.get('ChecksWarning'));
    // Otherwise, we need to filter the child checks by both failing
    // states
    } else {
      var checks = this.get('Checks');
      return (checks.filterBy('Status', 'critical').get('length') +
        checks.filterBy('Status', 'warning').get('length'));
    }
  }.property('Checks'),

  //
  // The number of passing checks within the service.
  //
  passingChecks: function() {
    // If the service was returned from `/v1/internal/ui/services`
    // then we have a aggregated value which we can just grab
    if (this.get('ChecksPassing') !== undefined) {
      return this.get('ChecksPassing');
    // Otherwise, we need to filter the child checks by both failing
    // states
    } else {
      return this.get('Checks').filterBy('Status', 'passing').get('length');
    }
  }.property('Checks'),

  //
  // The formatted message returned for the user which represents the
  // number of checks failing or passing. Returns `1 passing` or `2 failing`
  //
  checkMessage: function() {
    if (this.get('hasFailingChecks') === false) {
      return this.get('passingChecks') + ' passing';
    } else {
      return this.get('failingChecks') + ' failing';
    }
  }.property('Checks'),

  nodes: function() {
    return (this.get('Nodes'));
  }.property('Nodes'),

  //
  // Boolean of whether or not there are failing checks in the service.
  // This is used to set color backgrounds and so on.
  //
  hasFailingChecks: Ember.computed.gt('failingChecks', 0),

  //
  // Key used for filtering through an array of this model, i.e s
  // searching
  //
  filterKey: Ember.computed.alias('Name'),
});

//
// A Consul Node
//
App.Node = Ember.Object.extend({
  //
  // The number of failing checks within the service.
  //
  failingChecks: function() {
    return this.get('Checks').reduce(function(sum, check) {
      var status = Ember.get(check, 'Status');
      // We view both warning and critical as failing
      return (status === 'critical' || status === 'warning') ?
        sum + 1 :
        sum;
    }, 0);
  }.property('Checks'),

  //
  // The number of passing checks within the service.
  //
  passingChecks: function() {
    return this.get('Checks').filterBy('Status', 'passing').get('length');
  }.property('Checks'),

  //
  // The formatted message returned for the user which represents the
  // number of checks failing or passing. Returns `1 passing` or `2 failing`
  //
  checkMessage: function() {
    if (this.get('hasFailingChecks') === false) {
      return this.get('passingChecks') + ' passing';
    } else {
      return this.get('failingChecks') + ' failing';
    }
  }.property('Checks'),

  //
  // Boolean of whether or not there are failing checks in the service.
  // This is used to set color backgrounds and so on.
  //
  hasFailingChecks: Ember.computed.gt('failingChecks', 0),

  //
  // The number of services on the node
  //
  numServices: Ember.computed.alias('Services.length'),

  services: Ember.computed.alias('Services'),

  filterKey: Ember.computed.alias('Node')
});


//
// A key/value object
//
App.Key = Ember.Object.extend(Ember.Validations.Mixin, {
  // Validates using the Ember.Valdiations library
  validations: {
    Key: { presence: true }
  },

  // Boolean if the key is valid
  keyValid: Ember.computed.empty('errors.Key'),
  // Boolean if the value is valid
  valueValid: Ember.computed.empty('errors.Value'),

  // The key with the parent removed.
  // This is only for display purposes, and used for
  // showing the key name inside of a nested key.
  keyWithoutParent: function() {
    return (this.get('Key').replace(this.get('parentKey'), ''));
  }.property('Key'),

  // Boolean if the key is a "folder" or not, i.e is a nested key
  // that feels like a folder. Used for UI
  isFolder: function() {
    if (this.get('Key') === undefined) {
      return false;
    }
    return (this.get('Key').slice(-1) === '/');
  }.property('Key'),

  // Boolean if the key is locked or now
  isLocked: function() {
    if (!this.get('Session')) {
      return false;
    } else {
      return true;
    }
  }.property('Session'),

  // Determines what route to link to. If it's a folder,
  // it will link to kv.show. Otherwise, kv.edit
  linkToRoute: function() {
    if (this.get('Key').slice(-1) === '/') {
      return 'kv.show';
    } else {
      return 'kv.edit';
    }
  }.property('Key'),

  // The base64 decoded value of the key.
  // if you set on this key, it will update
  // the key.Value
  valueDecoded: function(key, value) {

    // setter
    if (arguments.length > 1) {
      this.set('Value', value);
      return value;
    }

    // getter

    // If the value is null, we don't
    // want to try and base64 decode it, so just return
    if (this.get('Value') === null) {
      return "";
    }
    if (Base64.extendString) {
      // you have to explicitly extend String.prototype
      Base64.extendString();
    }
    // base64 decode the value
    return (this.get('Value').fromBase64());
  }.property('Value'),


  // An array of the key broken up by the /
  keyParts: function() {
    var key = this.get('Key');

    // If the key is a folder, remove the last
    // slash to split properly
    if (key.slice(-1) == "/") {
      key = key.substring(0, key.length - 1);
    }

    return key.split('/');
  }.property('Key'),

  // The parent Key is the key one level above this.Key
  // key: baz/bar/foobar/
  // grandParent: baz/bar/
  parentKey: function() {
    var parts = this.get('keyParts').toArray();

    // Remove the last item, essentially going up a level
    // in hiearchy
    parts.pop();

    return parts.join("/") + "/";
  }.property('Key'),

  // The grandParent Key is the key two levels above this.Key
  // key: baz/bar/foobar/
  // grandParent: baz/
  grandParentKey: function() {
    var parts = this.get('keyParts').toArray();

    // Remove the last two items, jumping two levels back
    parts.pop();
    parts.pop();

    return parts.join("/") + "/";
  }.property('Key')
});

//
// An ACL
//
App.Acl = Ember.Object.extend({
  isNotAnon: function() {
    if (this.get('ID') === "anonymous"){
      return false;
    } else {
      return true;
    }
  }.property('ID')
});

// Wrap localstorage with an ember object
App.Settings = Ember.Object.extend({
  unknownProperty: function(key) {
    return localStorage[key];
  },

  setUnknownProperty: function(key, value) {
    if(Ember.isNone(value)) {
      delete localStorage[key];
    } else {
      localStorage[key] = value;
    }
    this.notifyPropertyChange(key);
    return value;
  },

  clear: function() {
    this.beginPropertyChanges();
    for (var i=0, l=localStorage.length; i<l; i++){
      this.set(localStorage.key(i));
    }
    localStorage.clear();
    this.endPropertyChanges();
  }
});


