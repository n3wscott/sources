/*
 * bear - index.js
 */

function receive(msg) {
  msg = JSON.parse(msg.data); // lol
  console.log("receive:", msg);
  fish.push({
	msg: msg,
	t: 0,
  });
}

function eat(f) {
  socket.send(JSON.stringify({
    "type": "eat",
    "to": f.msg.from,
    "nonce": f.msg.nonce,
  }));
  flash("You ate "+escape(f.msg.from.name)+", yum!");
}

function enter_game() {
  name = document.getElementById("name").value;
  console.log("enter game for name", name);
  connect(name, socket => {
    canvas_init(eat);
    socket.onmessage = receive;
    hide("entergame");
    show("game");
  });
}
