var conn = null

function createConn() {
    const socket = new WebSocket("ws://localhost:8080")
    socket.binaryType = "text"

    console.log(socket)

    socket.addEventListener("error", (event) => {
        console.error("Error while trying to connect to WebSocket")
        return
    })

    socket.addEventListener("message", (event) => {
        console.log(event)
    })

    socket.addEventListener("close", (event) => {
        console.log(event)
    })

    return socket
}

function sendMessage(message) {
    conn.send(JSON.stringify(message.toString()))
    console.log('sent: ' + message)
}

document.addEventListener("DOMContentLoaded", (event) => {
    conn = createConn()
})

document.addEventListener("submit", (event) => {
    event.preventDefault()

    const msg = document.getElementById("userMessage").value

    sendMessage(msg)
})
