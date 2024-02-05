(() => {
    document.body.classList.add('has-js')
    var menutoggle = document.getElementById('menutoggle')
    var shadow = document.getElementById('shadow')
    var menu = document.getElementById('menu')
    var nav = document.querySelector('nav')
    menutoggle.addEventListener('click', function () {
        menu.scrollTop = 0
        nav.classList.add('open')
    })
    shadow.addEventListener('click', function (e) {
        nav.classList.remove('open')
    })
})()