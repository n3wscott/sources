/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
    addpoint();
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
