"use strict"



function doClick() {
    alert("Hola mundo")
}

function doLogin() {
    fetch("/api/").then(resp => resp.text())
        .then(result => console.log("Result ", result))
        .catch(er => console.error(er))
}
