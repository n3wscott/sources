/*
 * bear - index.js
 */

function receive(msg) {
  msg = JSON.parse(msg.data); // lol
  console.log("receive:", msg);
  btn = document.getElementById("eat");
  fish = msg.from;
  nonce = msg.nonce;
  btn.innerHTML = "Eat "+escape(msg.from.name)+"!";
  show("eat");
  window.setTimeout(hide, 3 * 1000, "eat");
}

function eat() {
  socket.send(JSON.stringify({
    "type": "eat",
    "to": fish,
    "nonce": nonce,
  }));
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
