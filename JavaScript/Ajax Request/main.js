var ourRequest = new XMLHttpRequest();
ourRequest.open('GET', 'URL')

ourRequest.onload = function() {
    console.log(ourRequest.responseText);
}
ourRequest.send();