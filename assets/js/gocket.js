(function(win){

  win.gio = function(url){
    if (!url){
      url = "ws://" + location.host + "/sockets";
    }

    var wsocket = (WebSocket || MozWebSocket);
    var socket = new wsocket(url);
    var gocket = new Gocket(socket);
    return gocket;
  };


  var closed, errored, open, url;
  var closedCallbacks = [];
  var errorCallbacks = [];
  var messageCallbacks = [];

  function Gocket(socket){
    this.socket = socket;
    errored = closed = false
    open = true;

    this.socket.addEventListener("close", function(evt){
      closed = true;
      open = false;
      errored = !evt.wasClean;

      closedCallbacks.forEach(function(cb){ cb(evt.wasClean, evt.code, evt.reason); });
    });

    this.socket.addEventListener("message", function(evt){
      var payload = JSON.parse(evt.data);
      messageCallbacks.forEach(function(cb){ cb(payload); });
    });

    this.socket.addEventListener("error", function(){
      errored = true;
      closed = true;
      open = false;

      errorCallbacks.forEach(function(cb){ cb(); });
    });
  };

  Gocket.prototype.closed = function(callback){
    closedCallbacks.push(callback);
    return this;
  };

  Gocket.prototype.errored = function(callback){
    errorCallbacks.push(callback);
    return this;
  };

  Gocket.prototype.messaged = function(callback){
    messageCallbacks.push(callback);
    return this;
  };

  Gocket.prototype.send = function(item){
    if(closed || errored)
      throw new Error("This socket has closed");

    var payload = JSON.stringify(item);
    this.socket.send(payload);
  };

  Gocket.prototype.close = function(reason, shortCode){
    if (!shortCode && reason)
      shortCode = 1000;
   else 
      shortCode = reason = void 0; 

    this.socket.close(shortCode, reason);
  };
  
}(window))
