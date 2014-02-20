// Foundation JavaScript
// Documentation can be found at: http://foundation.zurb.com/docs
$(document).foundation();


$(function() {
    var ws = new WebSocket("ws://localhost:8080/monitor");
    ws.onmessage = function(event) {
        var obj = $.parseJSON(event.data);
        console.log("Received obj: ", obj);
        $("#data").text(obj.Name + " " + obj.Age);
        $("#data").addClass(obj.Status);
    };
});
