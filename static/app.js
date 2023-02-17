let username = null;
let input = null;
let output = null;
let scoreBoard = null
let socket = null;
  

function MessageAdd(message) {
	var chat_messages = document.getElementById("chat-messages");

	chat_messages.insertAdjacentHTML("beforeend", message);
	chat_messages.scrollTop = chat_messages.scrollHeight;
}

function promptForUsername() {
    while (!!!username) {
        username = prompt("Enter your name: ");
    }
}

function initialize() {
    promptForUsername();
    document.querySelector("#name").innerHTML =	username    
    // username = document.getElementById("name")
    input = document.getElementById("input");
    output = document.getElementById("output");
    scoreBoard = document.getElementById("score-board");
    let connectHost = location.hostname + ":" + location.port;
    socket = new WebSocket(`ws://${connectHost}/connect`);
    
    socket.onopen = function () {
        output.innerHTML += "Status: Connected\n";
    };
    
    socket.onmessage = function (e) {
        MessageAdd('<div class="message">' + e.data + '</div>')
    };
    
    input.addEventListener('keypress', e => {if (e.key === 'Enter') send()})
}

function send() {
    if(input.value === ""){
        return
    }
    fetch("/send-message", {
        method: "post",
        //make sure to serialize your JSON body
        body: JSON.stringify({
            Value: input.value,
            User: username,
        })
    })
    input.value = ""
}
function registerBtnHit(){
    fetch("/register-hit", {
        method: "post",
        body: JSON.stringify({
            user: username,
        })
    })
}

setInterval(()=> {
    fetch("/top-scorers").then(async (resp) => {
        let scoreBoardContent = ""
        let scores = await resp.json()
        scores?.elements.forEach((v) => {
            scoreBoardContent += `<h3>${v.name}: ${v.value}</h3>`;
        })
        scoreBoard.innerHTML = scoreBoardContent
    })
}, 1000)

window.onload = initialize