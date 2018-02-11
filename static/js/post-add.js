$("#button").on('click', function () {
    var $url = 'http://pluma.me/post/add?url=' + $("#inputUrl").val() + "&title=" + $("#inputTitle").val() + "#load-stuff";
    var xhr = new XMLHttpRequest();
    xhr.open("GET", $url, true);
    xhr.onload = function (e) {
        if (xhr.readyState === 4) {
            if (xhr.status === 200) {
                console.log(xhr.responseText);
            } else {
                console.error(xhr.statusText);
            }
        }
    };
    xhr.onerror = function (e) {
        console.error(xhr.statusText);
    };
    xhr.send(null);
    alert('success');
    window.location.replace("http://pluma.me/admin/post");
});
