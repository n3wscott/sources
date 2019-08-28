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

// shared javascript

function hide(id) {
  document.getElementById(id).style.display = "none";
}

function show(id) {
  document.getElementById(id).style.display = "inline";
}

// very awful. but it works.
function uuidv4() {
  return ([1e7]+-1e3+-4e3+-8e3+-1e11).replace(/[018]/g, c =>
      (c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16));
}

function connect(name, then) {
  name = name.replace(/ /g, '_');
  var openingsocket = new WebSocket("ws://"+window.location.host+"/websocket?name="+encodeURIComponent(name));
  openingsocket.onopen = function(_) {
    console.log("socket opened");
    // make it globally available
    socket = openingsocket;
    then(socket);
  }
}

function canvas_init(click_handle) {
  c = document.querySelector("#c");
  c.width = 480;
  c.height = 240;
  cx = c.getContext("2d");
  cx.font = "48px Arial";
  cx.textAlign = "center";
  time = 0;
  frame = 0;
  fish = [];

  c.addEventListener('click', fish_click(click_handle), false);

  loop();
}

// draw the fish forever
function loop() {
  requestAnimationFrame(loop);
  time = frame/60;
  frame++;
  cx.clearRect(0, 0, c.width, c.height);
  draw_fish(time);
}

// draws fish for time t
function draw_fish(t) {
  for (let f of fish) {
    if (!f.ot) {
      f.ot = t;
    }
    if (!f.ar) {
      // angle random -- start in a random rotation on unit circle
      f.ar = Math.random() * Math.PI * 2
    }
    f.t = t;
    dt = f.t - f.ot;

    cx.save();
    let x = (dt*60*3);
    let y = (1/300) * (x-240) ** 2 + 50;
    cx.translate(x, y);
    cx.rotate(f.ar + dt * Math.PI);
    cx.fillText("🐟", 0, 0);
    cx.restore();

    f.x = x;
    f.y = y;

    /*
    const [centerx, centery] = fish_center(f);
    cx.beginPath();
    cx.arc(centerx, centery, 20, 0, Math.PI*2);
    cx.stroke();
    */
  }

  // delete old fishies
  fish = fish.filter(f => (f.t - f.ot) <= 3 && (!f.delete))
}

// find the fish you clicked on and call the handle function
function fish_click(handle) {
  return function(ev) {
    var x = event.pageX - c.offsetLeft;
    var y = event.pageY - c.offsetTop;

    for (let f of fish) {
      if (!f.ot || !f.ar) {
        continue; // fish hasn't been drawn yet, also missing critical info
      }

      const [centerx, centery] = fish_center(f);

      if (((centerx - x)**2 + (centery - y)**2)**0.5 < 24) {
        handle(f);
        f.delete = true;
      }
    }
  }
}

// compute the center of a fish
function fish_center(f) {
      angle = (f.ar + (f.t - f.ot) * Math.PI + Math.PI/2);
      rcos = Math.cos(angle);
      rsin = Math.sin(angle);
      centerx = f.x - (24 * rcos); // font is 48 so half of it is 24
      centery = f.y - (24 * rsin);
      return [centerx, centery];
}

// display a message under the box
function flash(msg) {
  stat = document.getElementById("status");
  stat.innerHTML = "<span>"+msg+"</span><br>"+stat.innerHTML;
}

// give user a point
function addpoint() {
  pointbox = document.getElementById("points");
  pointbox.innerHTML = parseInt(pointbox.innerHTML)+1
}
