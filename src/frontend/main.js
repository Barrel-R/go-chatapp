var conn = null
var chat = null

function createConn() {
    const socket = new WebSocket("ws://localhost:8080")
    socket.binaryType = "text"

    console.log(socket)

    socket.addEventListener("error", (event) => {
        console.error("Error while trying to connect to WebSocket")
        return
    })

    socket.addEventListener("message", (event) => {
        addMessage(event.data)
    })

    socket.addEventListener("close", (event) => {
        console.log(event)
    })

    return socket
}

function addMessage(message, isUser = false) {
    const el = document.createElement('div')
    el.classList = "rounded w-fit bg-slate-600 p-2"

    if (isUser) {
        el.classList += " ms-auto"
    }

    el.innerText = message
    chat.append(el)

    if (chat.scrollHeight > 308) {
        chat.scrollTop = chat.scrollHeight
    }
}

function sendMessage(message) {
    conn.send(JSON.stringify(message.toString()))
    addMessage(message, true)
}

document.addEventListener("DOMContentLoaded", (event) => {
    conn = createConn()
    chat = document.getElementById("chat")
})

document.addEventListener("submit", (event) => {
    event.preventDefault()

    const msg = document.getElementById("userMessage").value

    sendMessage(msg)
})
