Ember.Handlebars.helper('panelBar', function(status) {
  var highlightClass;

  if (status == "passing") {
    highlightClass = "bg-green";
  } else {
    highlightClass = "bg-orange";
  }
  return new Handlebars.SafeString('<div class="panel-bar ' + highlightClass + '"></div>');
});

Ember.Handlebars.helper('listBar', function(status) {
  var highlightClass;

  if (status == "passing") {
    highlightClass = "bg-green";
  } else {
    highlightClass = "bg-orange";
  }
  return new Handlebars.SafeString('<div class="list-bar-horizontal ' + highlightClass + '"></div>');
});

Ember.Handlebars.helper('sessionName', function(session) {
  var name;

  if (session.Name === "") {
    name = '<span>' + Handlebars.Utils.escapeExpression(session.ID) + '</span>';
  } else {
    name = '<span>' + Handlebars.Utils.escapeExpression(session.Name) + '</span>' + ' <small>' + Handlebars.Utils.escapeExpression(session.ID) + '</small>';
  }

  return new Handlebars.SafeString(name);
});

Ember.Handlebars.helper('sessionMeta', function(session) {
  var meta = '<div class="metadata">' + Handlebars.Utils.escapeExpression(session.Behavior) + ' behavior</div>';

  if (session.TTL !== "") {
    meta = meta + '<div class="metadata">, ' + Handlebars.Utils.escapeExpression(session.TTL) + ' TTL</div>';
  }

  return new Handlebars.SafeString(meta);
});

Ember.Handlebars.helper('aclName', function(name, id) {
  if (name === "") {
    return id;
  } else {
    return new Handlebars.SafeString(Handlebars.Utils.escapeExpression(name) + ' <small class="pull-right no-case">' + Handlebars.Utils.escapeExpression(id) + '</small>');
  }
});


Ember.Handlebars.helper('formatRules', function(rules) {
  if (rules === "") {
    return "No rules defined";
  } else {
    return rules;
  }
});


// We need to do this because of our global namespace properties. The
// service.Tags
Ember.Handlebars.helper('serviceTagMessage', function(tags) {
  if (tags === null) {
    return "No tags";
  }
});


// Sends a new notification to the UI
function notify(message, ttl) {
  if (window.notifications !== undefined && window.notifications.length > 0) {
    $(window.notifications).each(function(i, v) {
      v.dismiss();
    });
  }
  var notification = new NotificationFx({
    message : '<p>'+ message + '</p>',
    layout : 'growl',
    effect : 'slide',
    type : 'notice',
    ttl: ttl,
  });

  // show the notification
  notification.show();

  // Add the notification to the queue to be closed
  window.notifications = [];
  window.notifications.push(notification);
}

// Tomography

// TODO: not sure how to how do to this more Ember.js-y
function tomographyMouseOver(el) {
  var buf = el.getAttribute('data-node') + ' - ' + el.getAttribute('data-distance') + 'ms';
  document.getElementById('tomography-node-info').innerHTML = buf;
}

Ember.Handlebars.helper('tomographyGraph', function(tomography, size) {

  // This is ugly, but I'm working around bugs with Handlebars and templating
  // parts of svgs. Basically things render correctly the first time, but when
  // stuff is updated for subsequent go arounds the templated parts don't show.
  // It appears (based on google searches) that the replaced elements aren't
  // being interpreted as http://www.w3.org/2000/svg. Anyway, this works and
  // if/when Handlebars fixes the underlying issues all of this can be cleaned
  // up drastically.

  var max = -999999999;
  tomography.distances.forEach(function (d, i) {
    if (d.distance > max) {
      max = d.distance;
    }
  });
  var insetSize = size / 2 - 8;
  var buf = '' +
'      <svg width="' + size + '" height="' + size + '">' +
'        <g class="tomography" transform="translate(' + (size / 2) + ', ' + (size / 2) + ')">' +
'          <g>' +
'            <circle class="background" r="' + insetSize + '"/>' +
'            <circle class="axis" r="' + (insetSize * 0.25) + '"/>' +
'            <circle class="axis" r="' + (insetSize * 0.5) + '"/>' +
'            <circle class="axis" r="' + (insetSize * 0.75) + '"/>' +
'            <circle class="border" r="' + insetSize + '"/>' +
'          </g>' +
'          <g class="lines">';
  var distances = tomography.distances;
  var n = distances.length;
  if (tomography.n > 360) {
    // We have more nodes than we want to show, take a random sampling to keep
    // the number around 360.
    var sampling = 360 / tomography.n;
    distances = distances.filter(function (_, i) {
      return i == 0 || i == n - 1 || Math.random() < sampling
    });
    // Re-set n to the filtered size
    n = distances.length;
  }
  distances.forEach(function (d, i) {
    buf += '            <line transform="rotate(' + (i * 360 / n) + ')" y2="' + (-insetSize * (d.distance / max)) + '" ' +
      'data-node="' + d.node + '" data-distance="' + d.distance + '" onmouseover="tomographyMouseOver(this);"/>';
  });
  buf += '' +
'          </g>' +
'          <g class="labels">' +
'            <circle class="point" r="5"/>' +
'            <g class="tick" transform="translate(0, ' + (insetSize * -0.25 ) + ')">' +
'              <line x2="70"/>' +
'              <text x="75" y="0" dy=".32em">' + (max > 0 ? (parseInt(max * 25) / 100) : 0) + 'ms</text>' +
'            </g>' +
'            <g class="tick" transform="translate(0, ' + (insetSize * -0.5 ) + ')">' +
'              <line x2="70"/>' +
'              <text x="75" y="0" dy=".32em">' + (max > 0 ? (parseInt(max * 50) / 100) : 0)+ 'ms</text>' +
'            </g>' +
'            <g class="tick" transform="translate(0, ' + (insetSize * -0.75 ) + ')">' +
'              <line x2="70"/>' +
'              <text x="75" y="0" dy=".32em">' + (max > 0 ? (parseInt(max * 75) / 100) : 0) + 'ms</text>' +
'            </g>' +
'            <g class="tick" transform="translate(0, ' + (insetSize * -1) + ')">' +
'              <line x2="70"/>' +
'              <text x="75" y="0" dy=".32em">' + (max > 0 ? (parseInt(max * 100) / 100) : 0) + 'ms</text>' +
'            </g>' +
'          </g>' +
'        </g>' +
'      </svg>';

  return new Handlebars.SafeString(buf);
});
