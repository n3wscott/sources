/*
 * salmon - index.js
 */

function receive(msg) {
  msg = JSON.parse(msg.data); // lol
  console.log("receive:", msg);
  fish = fish.filter(f => f.nonce != msg.nonce);
  if (msg.type == "eat") {
    flash("You were eaten by "+escape(msg.from.name)+".");
  }
  else {
	flash("The bear named "+escape(msg.from.name)+" starves tonight.");
  }
}

function enter_game() {
  name = document.getElementById("name").value;
  console.log("enter game for name", name);
  connect(name, socket => {
    canvas_init();
    socket.onmessage = receive;
    hide("entergame");
    show("game");
  });
}

function jump() {
  console.log("sending jump");
  let nonce = uuidv4();
  socket.send(JSON.stringify({
    "nonce": nonce,
  }));
  fish.push({
    nonce: nonce,
    t: 0,
  });
}
