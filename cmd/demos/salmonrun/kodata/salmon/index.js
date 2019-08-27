/*
 * salmon - index.js
 */

function receive(msg) {
  msg = JSON.parse(msg.data); // lol
  console.log("receive:", msg);
  stat = document.getElementById("status");
  if (msg.type == "eat") {
    stat.innerHTML = "<span>You were eaten by "+escape(msg.from.name)+".</span><br>"+stat.innerHTML;
    fish.filter(f => f.nonce != msg.nonce)
  }
  else {
    stat.innerHTML = "<span>The bear named "+escape(msg.from.name)+" starves tonight.</span><br>"+stat.innerHTML;
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
