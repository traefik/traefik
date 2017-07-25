var Serf = (function() {

  function initialize (){
    Serf.Util.runIfClassNamePresent('page-home', initHome);
    
    new Sidebar();
  }

  function initHome() {
    if(!Serf.Util.isMobile){
      Serf.Nodes.init();
    }else{
      Serf.Home.mobileHero();
    }
  }

    //api
  return {
    initialize: initialize
    }

})();

var Serf = Serf || {};

(function () {

  //check for mobile user agents
  var isMobile = (function(){
     if( navigator.userAgent.match(/Android/i)
     || navigator.userAgent.match(/webOS/i)
     || navigator.userAgent.match(/iPhone/i)
     //|| navigator.userAgent.match(/iPad/i)
     || navigator.userAgent.match(/iPod/i)
     || navigator.userAgent.match(/BlackBerry/i)
     || navigator.userAgent.match(/Windows Phone/i)
     ){
      return true;
      }
     else {
        return false;
      }
    })()

    // calls the given function if the given classname is found
    function runIfClassNamePresent(selector, initFunction) {
        var elms = document.getElementsByClassName(selector);
        if (elms.length > 0) {
            initFunction();
        }
    }

    Serf.Util = {};
    Serf.Util.isMobile = isMobile;
    Serf.Util.runIfClassNamePresent = runIfClassNamePresent;

})();

var Serf = Serf || {};

(function () {

    // calls the given function if the given classname is found
    function mobileHero() {
      var jumbo = document.getElementById('jumbotron');
      jumbo.className = jumbo.className + ' mobile-hero';
    }

    Serf.Home = {};
    Serf.Home.mobileHero = mobileHero;

})();

var Serf = Serf || {};

(function () {

    var width = 1400,
        height = 490,
    border = 50,
        numberNodes = 128,
        linkGroup = 0;
        //nodeLinks = [];

  var nodes = [];
  for (i=0; i<numberNodes; i++) {
    nodes.push({
      x: Math.random() * (width - border) + (border / 2),
      y: Math.random() * (height - border) + (border / 2),
    });
  }

    var fill = d3.scale.category20();

    var force = d3.layout.force()
    .size([width, height])
        .nodes(nodes)
      .linkDistance(60)
    .charge(-1)
    .gravity(0.0004)
        .on("tick", tick);

    var svg = d3.select("#jumbotron").append("svg")
        .attr('id', 'node-canvas')
        .attr("width", width)
        .attr("height", height)

    //set left value after adding to dom
    resize();

    svg.append("rect")
        .attr("width", width)
        .attr("height", height);

    var nodes = force.nodes(),
        links = force.links(),
        node = svg.selectAll(".node"),
        link = svg.selectAll(".link");

    var cursor = svg.append("circle")
        .attr("r", 30)
        .attr("transform", "translate(-100,-100)")
        .attr("class", "cursor");


    function createLink(index) {
        var node = nodes[index];
        var nodeSelected = svg.select("#id_" + node.index).classed("active linkgroup_"+ linkGroup, true);

    var distMap = {};
    var distances = [];

    for (var i=0; i<nodes.length; i++) {
      if (i == index) {
        continue
      }

      var target = nodes[i];
            var selected = svg.select("#id_" + i);
            var dx = selected.attr('cx') - nodeSelected.attr('cx');
            var dy = selected.attr('cy') - nodeSelected.attr('cy');
      var dist = Math.sqrt(dx * dx + dy * dy)

      if (dist in distMap) {
        distMap[dist].push(target)
      } else {
        distMap[dist] = [target]
      }
      distances.push(dist)
    }

    distances.sort(d3.ascending);
    for (i = 0; i < 3; i++) {
      var dist = distances[i]
      var target = distMap[dist].pop()
      var link  = {
        source: node,
        target: target
      }
      links.push(link);
    }

        restart();
    }


    function tick() {
    link.attr("x1", function(d) { return d.source.x; })
        .attr("y1", function(d) { return d.source.y; })
        .attr("x2", function(d) { return d.target.x; })
        .attr("y2", function(d) { return d.target.y; });

    node.attr("cx", function(d) { return d.x; })
        .attr("cy", function(d) { return d.y; });
    }


    function restart() {

        node = node.data(nodes);

        node.enter().insert("circle", ".cursor")
            .attr("class", "node")
            .attr("r", 12)
            .attr("id", function (d, i) {
                return ("id_" + i)
            })
            .call(force.drag);

        link = link.data(links);

        link.enter().insert("line", ".node")
            .attr("class", "link active linkgroup_"+ linkGroup);

        force.start();

        resetLink(linkGroup);
        linkGroup++;
    }

    function resetLink(num){
      setTimeout(resetColors, 700, num)
    }

    function resetColors(num){
    svg.selectAll(".linkgroup_"+ num).classed('active', false)
    }

    window.onresize = function(){
        resize();
    }

    function resize() {
      var nodeC = document.getElementById('node-canvas');
        wW = window.innerWidth;

      nodeC.style.left = ((wW - width) / 2 ) + 'px';
    }

    //kick things off
    function init() {
      restart();
    for (i=0;i<numberNodes;i++) {
      setTimeout(createLink, 700*i+1000, i);
    }
    }

  Serf.Nodes = {};
    Serf.Nodes.init = init;

})();
