let btn = document.getElementById('btn') 

btn.addEventListener('click', function() {
    var ourRequest = new XMLHttpRequest();
ourRequest.open('GET', 'URL')

ourRequest.onload = function() {
    var ourData = JSON.parse(ourRequest.responseText)
    console.log(ourRequest[1])
    
}
ourRequest.send();

})




