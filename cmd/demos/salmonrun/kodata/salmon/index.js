/*
 * salmon - index.js
 */

function hide(id) {
	document.getElementById(id).style.display = "none"
}

function show(id) {
	document.getElementById(id).style.display = "inline"
}

function connect(name, then) {
	var openingsocket = new WebSocket("ws://"+window.location.host+"/websocket?name="+encodeURIComponent(name))
	openingsocket.onopen = function(_) {
		console.log("socket opened")
		// make it globally available
		socket = openingsocket
		then(socket)
	}
	openingsocket.onmessage = receive
}

function receive(msg) {
	msg = JSON.parse(msg.data) // lol
	console.log("receive:", msg)
	stat = document.getElementById("status")
	if (msg.type == "eat") {
		stat.innerHTML = "<span>You were eaten by "+escape(msg.from.name)+"</span><br>"+stat.innerHTML
	}
	else {
		stat.innerHTML = "<span>The bear named "+escape(msg.from.name)+" starves tonight.</span><br>"+stat.innerHTML
	}
}

function enter_game() {
	name = document.getElementById("name").value
	console.log("enter game for name", name)
	connect(name, socket => {
		hide("entergame")
		show("game")
	})
}

function jump() {
	console.log("sending jump")
	socket.send("jump")
}
