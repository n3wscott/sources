/*
 * bear - index.js
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
	btn = document.getElementById("eat")
	fish = msg.from
	nonce = msg.nonce
	btn.innerHTML = "Eat "+escape(msg.from.name)+"!"
	show("eat")
	window.setTimeout(hide, 3 * 1000, "eat");
}

function eat() {
	socket.send(JSON.stringify({
		"type": "eat",
		"to": fish,
		"nonce": nonce,
	}))
}

function enter_game() {
	name = document.getElementById("name").value
	console.log("enter game for name", name)
	connect(name, socket => {
		hide("entergame")
		show("game")
	})
}
