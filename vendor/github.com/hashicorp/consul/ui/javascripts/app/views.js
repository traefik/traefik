
//
// DC
//

App.DcView = Ember.View.extend({
    templateName: 'dc',
    classNames: 'dropdowns',

    click: function(e){
        if ($(e.target).is('.dropdowns')){
          $('ul.dropdown-menu').hide();
        }
    }
});


App.ItemView = Ember.View.extend({
    templateName: 'item'
});

//
// Services
//
App.ServicesView = Ember.View.extend({
    templateName: 'services',
});

App.ServicesShowView = Ember.View.extend({
    templateName: 'service'
});

App.ServicesLoadingView = Ember.View.extend({
    templateName: 'item/loading'
});

//
// Nodes
//

App.NodesView = Ember.View.extend({
    templateName: 'nodes'
});

App.NodesShowView = Ember.View.extend({
    templateName: 'node'
});

App.NodesLoadingView = Ember.View.extend({
    templateName: 'item/loading'
});


// KV

App.KvListView = Ember.View.extend({
    templateName: 'kv'
});

// Actions

App.ActionBarView = Ember.View.extend({
    templateName: 'actionbar'
});

// ACLS

App.AclView = Ember.View.extend({
    templateName: 'acls',
});

App.AclsShowView = Ember.View.extend({
    templateName: 'acl'
});


// Settings

App.SettingsView = Ember.View.extend({
    templateName: 'settings',
});
