function hide(id) {
  document.getElementById(id).style.display = "none"
}

function show(id) {
  document.getElementById(id).style.display = "inline"
}

function connect(name) {
  var openingsocket = new WebSocket("wss://"+window.location.host+"/websocket?name="+encodeURIComponent(name))
  openingsocket.onopen = function(_) {
    console.log("socket opened")
        // make it globally available
        socket = openingsocket
  }
}

function enter_game() {
  name = document.getElementById("name").value
  console.log("enter game for name", name)
  connect(name)
  hide("entergame")
  show("game")
}

function jump() {
  socket.send("jump")
}
