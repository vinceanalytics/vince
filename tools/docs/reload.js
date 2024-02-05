const socket = new WebSocket("ws://" + window.location.host + "/reload");
socket.addEventListener("message", () => {
    window.location.reload()
})
