function graph_settings() {
	return {
		chart: {
			type: 'spline'
		},
		title: {
			text: null
		},
		xAxis: {
			title: {
				text: "Time (sec)",
			}
		},
		yAxis: {
			title: {
				text: 'Convergence %'
			},
			min: 0.0,
			max: 100.0
		},
		tooltip: {
			formatter: function() {
				return '<b>'+ (Math.round(this.y*1000.0)/1000.0) +'%</b><br/>'
			}
		},
		legend: {
			enabled: false
		},
		series: [{
			name: 'Convergence Rate',
			data: []
		}]
	}
}

function create_graph() {
	$('#graph').highcharts(graph_settings())
	return $('#graph').highcharts()
}

var Simulator = Class.$extend({
	__init__ : function(graph, bytes, maxConverge) {
		this.graph = graph
		this.bytes = bytes
		this.maxConverge = maxConverge
		this.interval = 0.2
		this.fanout = 3
		this.nodes = 30
		this.packetLoss = 0
		this.nodeFail = 0
	},

	convergenceAtRound: function(x) {
		var contact = (this.fanout / this.nodes) * (1 - this.packetLoss) * (1 - this.nodeFail) * 0.5
		var numInf = this.nodes / (1 + (this.nodes+1) * Math.pow(Math.E, -1*contact*this.nodes*x))
		return numInf / this.nodes
	},

	roundLength: function() {
		return this.interval
	},

	seriesData: function() {
		var data = []
		var lastVal = 0
		var round = 0
		var roundLength = this.roundLength()
		while (lastVal < this.maxConverge && round < 100) {
			lastVal = this.convergenceAtRound(round)
			data.push([round * roundLength, lastVal*100.0])
			round++
		}
			return data
	},

	bytesUsed: function() {
		var roundLength = this.roundLength()
		var roundsPerSec = 1 / roundLength
		var packetSize = 1400
		var send = packetSize * this.fanout * roundsPerSec
		return send * 2
	},

	draw: function() {
		var data = this.seriesData()
		this.graph.series[0].setData(data, false)
		this.graph.redraw()

		var kilobits = this.bytesUsed() * 8
		var used = Math.round((kilobits / 1024) * 10) / 10
		this.bytes.html(""+used)
	}
})

function update_interval(elem, simulator) {
	var val = elem.value
	var interval = Number(val)
	if (isNaN(interval)) {
		alert("Gossip interval must be a number!")
		return
	}
	if (interval <= 0) {
		alert("Gossip interval must be a positive value!")
		return
	}
	simulator.interval = interval
	simulator.draw()
	console.log("Redraw with interval set to: " + interval)
}


function update_fanout(elem, simulator) {
	var val = elem.value
	var fanout = Number(val)
	if (isNaN(fanout)) {
		alert("Gossip fanout must be a number!")
		return
	}
	if (fanout <= 0) {
		alert("Gossip fanout must be a positive value!")
		return
	}
	simulator.fanout = fanout
	simulator.draw()
	console.log("Redraw with fanout set to: " + fanout)
}

function update_nodes(elem, simulator) {
	var val = elem.value
	var nodes = Number(val)
	if (isNaN(nodes)) {
		alert("Node count must be a number!")
		return
	}
	if (nodes <= 1) {
		alert("Must have at least one node")
		return
	}
	simulator.nodes = nodes
	simulator.draw()
	console.log("Redraw with nodes set to: " + nodes)
}

function update_packetloss(elem, simulator) {
	var val = elem.value
	var pkt = Number(val)
	if (isNaN(pkt)) {
		alert("Packet loss must be a number!")
		return
	}
	if (pkt < 0 || pkt >= 100) {
		alert("Packet loss must be greater or equal to 0 and less than 100")
		return
	}
	simulator.packetLoss = (pkt / 100.0)
	simulator.draw()
	console.log("Redraw with packet loss set to: " + pkt)
}

function update_failed(elem, simulator) {
	var val = elem.value
	var failed = Number(val)
	if (isNaN(failed)) {
		alert("Failure rate must be a number!")
		return
	}
	if (failed < 0 || failed >= 100) {
		alert("Failure rate must be greater or equal to 0 and less than 100")
		return
	}
	simulator.nodeFail = (failed / 100.0)
	simulator.draw()
	console.log("Redraw with failure rate set to: " + failed)
}

// wait for dom ready
$(function(){
	var bytes = $("#bytes")
	var graph = create_graph()
	var simulator = new Simulator(graph, bytes, 0.9999)
	simulator.draw()

	var interval = $("#interval")
	interval.change(function() { update_interval(interval[0], simulator) })

	var fanout = $("#fanout")
	fanout.change(function() { update_fanout(fanout[0], simulator) })

	var nodes = $("#nodes")
	nodes.change(function() { update_nodes(nodes[0], simulator) })

	var loss = $("#packetloss")
	loss.change(function() { update_packetloss(loss[0], simulator) })

	var failed = $("#failed")
	failed.change(function() { update_failed(failed[0], simulator) })
})
