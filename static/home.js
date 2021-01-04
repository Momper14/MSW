window.onload = function () {
    var conn;
    var command = document.getElementById("command");
    var log = document.getElementById("log");

    log.scrollTo(0, log.scrollHeight);

    function appendLog(item) {
        var doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
        log.appendChild(item);
        if (doScroll) {
            log.scrollTop = log.scrollHeight - log.clientHeight;
        }
    }


    function send(message) {
        if (!conn || !message) {
            return false;
        }
        conn.send(message);
    }

    document.getElementById("start").onclick = function () {
        send(JSON.stringify({
            target: "WRAPPER",
            payload: "start"
        }))
    };

    document.getElementById("restart").onclick = function () {
        send(JSON.stringify({
            target: "WRAPPER",
            payload: "restart"
        }))
    };

    document.getElementById("stop").onclick = function () {
        send(JSON.stringify({
            target: "WRAPPER",
            payload: "stop"
        }))
    };

    document.getElementById("form").onsubmit = function () {
        send(JSON.stringify({
            target: "SERVER",
            payload: command.value
        }));
        command.value = "";
        return false;
    };

    if (window["WebSocket"]) {
        let proto = "ws:"
        if (document.location.protocol == "https:") {
            proto = "wss:"
        }


        conn = new WebSocket(proto + "//" + document.location.host + document.location.pathname + "ws");
        conn.onclose = function (evt) {
            let item = document.createElement("div");
            item.classList.add("error");
            item.innerHTML = "<b>Connection closed.</b>";
            let elemStatus = document.getElementById("status");
            let cl = elemStatus.classList;
            elemStatus.value = "not connected";
            cl.remove("starting");
            cl.remove("online");
            cl.add("offline");
            appendLog(item);
        };
        conn.onmessage = function (evt) {
            let messages = evt.data.split('\n');
            for (let i = 0; i < messages.length; i++) {
                let msg = JSON.parse(messages[i]);

                switch (msg.type) {
                    case "LOG": {
                        let item = document.createElement("div");
                        item.innerText = msg.payload;
                        appendLog(item);
                        break
                    }
                    case "ERROR": {
                        let item = document.createElement("div");
                        item.classList.add("error");
                        item.innerText = msg.payload;
                        appendLog(item);
                        break
                    }
                    case "STATE": {
                        let elemStatus = document.getElementById("status");
                        let cl = elemStatus.classList;
                        elemStatus.value = msg.payload;
                        switch (msg.payload) {
                            case "starting":
                                cl.remove("offline");
                                cl.remove("starting");
                                cl.add("starting");
                                break;
                            case "online":
                                cl.remove("offline");
                                cl.remove("starting");
                                cl.add("online");
                                break;
                            case "stopping":
                            case "offline":
                                cl.remove("starting");
                                cl.remove("online");
                                cl.add("offline");
                                break;
                        }
                        break
                    }
                    default:
                        console.log("unknown type " + msg.type);
                }
            }

        };
    } else {
        let item = document.createElement("div");
        item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
        appendLog(item);
    }
};