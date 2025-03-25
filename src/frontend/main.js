const BASE_RECONNECT_TIMEOUT = 125
var conn = null
var chat = null
var reconnectTimeout = BASE_RECONNECT_TIMEOUT

function createConn() {
    const socket = new WebSocket("ws://localhost:8080")

    subscribe()
    setupListeners(socket)

    conn = socket

    return socket
}

function subscribe() {
    fetch("http://localhost:8080/subscribe")
        .then(response => {
            console.log(response)
        })
        .catch(err => {
            console.log(err)
        })
}

function setupListeners(socket) {
    socket.addEventListener("open", (event) => {
        console.log("Connected to Websocket.", event)
        socket.send(JSON.stringify("Connected to Websocket."))
        reconnectTimeout = BASE_RECONNECT_TIMEOUT
    })

    socket.addEventListener("message", (event) => {
        try {
            addMessage({ message: event.data, timestamp: new Date() })
        } catch (err) {
            console.error("Error while parsing server message: ", err)
        }
    })

    socket.addEventListener("close", (event) => {
        // Before first connect have a reconnect timeout of 250, doubling at each
        // reconnect attempt, until the limit of 10.000
        const timeout = Math.min(10000, reconnectTimeout += reconnectTimeout)
        console.log("Trying to reconnect to WebSocket with a timeout of: ", timeout)
        setTimeout(createConn, timeout)
    })

    socket.addEventListener("error", (event) => {
        console.error("Socket encountered error: ", event, 'Closing socket')
        socket.close()
    })
}

function parseGoDate(date) {
    return new Date(date).toLocaleString()
}

function createClockSVG() {
    const svgElement = document.createElementNS('http://www.w3.org/2000/svg', 'svg')
    svgElement.classList = "w-4 h-4"
    svgElement.setAttribute('viewBox', '0 0 24 24')
    svgElement.setAttribute('fill', 'none')
    svgElement.setAttribute('xmlns', 'http://www.w3.org/2000/svg')

    const pathElement = document.createElementNS('http://www.w3.org/2000/svg', 'path')
    pathElement.setAttribute('d', 'M12 7V12L14.5 10.5M21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3C16.9706 3 21 7.02944 21 12Z')
    pathElement.setAttribute('stroke', '#BDBDBD')
    pathElement.setAttribute('stroke-width', '2')
    pathElement.setAttribute('stroke-linecap', 'round')
    pathElement.setAttribute('stroke-linejoin', 'round')

    svgElement.appendChild(pathElement);

    return svgElement
}

function addMessage(messageObj, isUser = false) {
    let msg = messageObj.message

    const msgNode = document.createElement('div')
    const dateNode = document.createElement('span')
    const clockSvg = createClockSVG()

    msgNode.classList = "flex flex-col gap-y-1 rounded w-fit bg-slate-600 p-2 mb-1"
    dateNode.classList = "ms-auto text-gray-400 text-sm flex gap-x-0.5 items-center"

    if (isUser) {
        msgNode.classList += " ms-auto"
    }

    msgNode.innerText = msg
    dateNode.appendChild(clockSvg)

    const dateText = document.createTextNode(' ' + parseGoDate(messageObj.timestamp))

    dateNode.appendChild(dateText)
    msgNode.appendChild(dateNode)

    chat.append(msgNode)

    chat.scrollTop = chat.scrollHeight
}

function sendMessage(message) {
    const xhr = new XMLHttpRequest()
    xhr.open("POST", "http://localhost:8080/publish")
    xhr.setRequestHeader("Content-Type", "application/json")

    try {
        xhr.send(JSON.stringify(message.toString()))
        addMessage({ message, timestamp: new Date() }, true)
    } catch (err) {
        console.error(err)
    }
}

document.addEventListener("DOMContentLoaded", (event) => {
    createConn()

    chat = document.getElementById("chat")
})

document.addEventListener("submit", (event) => {
    event.preventDefault()

    if (conn.readyState == 2 || conn.readyState == 3) {
        console.log("Couldn't send message, connection is closed", conn)
        return
    }

    const msg = document.getElementById("userMessage").value

    sendMessage(msg)
})
