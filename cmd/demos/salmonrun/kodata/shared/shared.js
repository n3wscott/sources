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
  var openingsocket = new WebSocket("ws://"+window.location.host+"/websocket?name="+encodeURIComponent(name));
      openingsocket.onopen = function(_) {
        console.log("socket opened");
            // make it globally available
            socket = openingsocket;
            then(socket);
      }
}

function canvas_init() {
  c = document.querySelector("#c");
  c.width = 480;
  c.height = 240;
  cx = c.getContext("2d");
  cx.font = "48px Arial";
  time = 0;
  frame = 0;
  fish = [];

  loop();
}

function loop() {
  requestAnimationFrame(loop);
  time = frame/60;
  frame++;
  cx.clearRect(0, 0, c.width, c.height);
  u(time);
}

function u(t) {
  fish = fish.filter(f => f.t <= 3*60)

  for (let f of fish) {
    if (!f.ot) {
      f.ot = t
    }
    f.t = t;
    dt = f.t - f.ot;

    cx.save();
    let x = (dt*60*3)%480;
    let y = (1/300) * (x-240) ** 2 + 50;
    cx.translate(x, y);
    cx.rotate(t*Math.PI);
    cx.fillText("🐟", 0, 0);
    cx.strokeRect(0, 0, 48, -48);
    cx.restore();
  }
}

