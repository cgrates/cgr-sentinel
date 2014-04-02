// Foundation JavaScript
// Documentation can be found at: http://foundation.zurb.com/docs
$(document).foundation();


$(function() {
    var ws = new WebSocket("ws://" + location.hostname + ":3000/monitor");
    ws.onmessage = function(event) {
        $("#data").html(event.data);
    };
});
