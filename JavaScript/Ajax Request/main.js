let btn = document.getElementById('ID') 

btn.addEventListener('click', function() {
    var ourRequest = new XMLHttpRequest();
ourRequest.open('GET', 'URL')
ourRequest.onload = function() {
    var ourData = JSON.parse(ourRequest.responseText)
    renderHTML(ourData);
}
ourRequest.send();

})

function renderHTML (data) {
    
}



