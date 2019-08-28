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

function loop() {
  requestAnimationFrame(loop);
  time = frame/60;
  frame++;
  cx.clearRect(0, 0, c.width, c.height);
  u(time);
}

function u(t) {
  for (let f of fish) {
    if (!f.ot) {
      f.ot = t
    }
    f.t = t;
    dt = f.t - f.ot;

    cx.save();
    let x = (dt*60*3);
    let y = (1/300) * (x-240) ** 2 + 50;
    cx.translate(x, y);
    cx.rotate(f.ot + dt * Math.PI);
    cx.fillText("🐟", 0, 0);
    //cx.strokeRect(-24, 0, 48, -48);
    cx.restore();

	f.x = x;
	f.y = y;
  }

  // delete old fishies
  fish = fish.filter(f => (f.t - f.ot) <= 3 && (!f.delete))
}

function fish_click(handle) {
  return function(ev) {
	var x = event.pageX - c.offsetLeft;
	var y = event.pageY - c.offsetTop;

	for (let f of fish) {
	  if (!f.ot) {
		continue; // fish hasn't been drawn yet, also missing critical info
	  }
	  angle = (f.ot + (f.t - f.ot) * Math.PI);
	  rcos = Math.cos(angle);
	  rsin = Math.sin(angle);
	  centerx = f.x + 24 * rcos; // font is 48 so half of it is 24
	  centery = f.y + 24 * rsin;

	  console.log(centerx, centery)
	  console.log("your click: ", x, y)

	  if (((centerx - x)**2 + (centery - y)**2)**0.5 < 24) {
		handle(f);
		f.delete = true;
	  }
	}
  }
}

function flash(msg) {
  stat = document.getElementById("status");
  stat.innerHTML = "<span>"+msg+"</span><br>"+stat.innerHTML;
}

function addpoint() {
	pointbox = document.getElementById("points")
	pointbox.innerHTML = parseInt(pointbox.innerHTML)+1
}
