(function(){
  function initBoard(id){
    var target = $(id);
    target.empty()
    target.append("<div id='board'><ul></div>");

    var list = target.find("ul");
    for (var x = 0; x < 9; x++) {
        var element = $("<li ><div id='" + x + "'></li>");
       list.append(element);
    }

    var clickCallbacks = [];
    $(id + " #board ul").on("click", function boardClickHandler($evt){
      var elementClicked = $evt.target;
      if (elementClicked && elementClicked.id) {
        var squareClicked = elementClicked.id;
        var squareX = squareClicked % 3;
        var squareY = Math.floor(squareClicked / 3);
        clickCallbacks.forEach(function(cb){ cb(squareX, squareY); });
      }
    });

    return {
      onClick: function(cb){
        clickCallbacks.push(cb);
      }
    };
  }

 var showToast = false;
  Notification.requestPermission(function(perm) {
    if (perm === "granted"){
        showToast = true;
    }
  });
  

  function writeStatus(msg){
    if (showToast && msg.sender === "Opponent" || msg.sender === "Server"){
      var notification = new Notification(msg.sender, {
        body: msg.message
      });
      setTimeout(function(){notification.close();}, 3E3);
    } 
    var updates = $("#updates ul");
    updates.append("<li>" + (new Date()).toLocaleTimeString() + " - " + msg.sender + ": " + msg.message + "</li>");
  };

  function sendMessage(msg){
    socket.send({type: -15, message: msg });
  }

  function applyMoveToBoard(isX, x, y) {
     var id = x + y * 3;
      var moveChar  = isX ? "X" : "O";
     $("#"+id).html(moveChar)
  }

  var div = $('#updates ul');
  setInterval(function(){
      var pos = div.scrollTop();
      div.scrollTop(pos + div.outerHeight());
  }, 400)

  var socket = gio(),
  boardHandle, ourTurn,

  /*
   * four states
   * 1. negotiation
   * 2. our turn
   * 3. their turn
   * 4. over
   */
  stateMachine = [
    function negotiation(){
      boardHandle = initBoard("#target");
      boardHandle.onClick(function(x, y){
        if (ourTurn) {
          socket.send({type: -2, "x":x, "y":y})
          writeStatus({ sender: "You", message: "Sending move (" + x + ", " + y + ")" })
        }
      });
    },
    function startOurTurn(){
      $("#target").prop('disabled',false);
      ourTurn = true;
    },
    function startOpponentTurn(){
      $("#target").prop('disabled',true);
      ourTurn = false;
    },
    function endGame(state){
      $("#target").prop('disabled',false);
    }
  ];

  socket.messaged(function(msg){
     var type = msg.type;
     writeStatus(msg);
     if (type >= 0 && type < stateMachine.length)
       stateMachine[type](msg);
    else if (type == -5)
      applyMoveToBoard(msg.isX, msg.x, msg.y)
  });

}());
