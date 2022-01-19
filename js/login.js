function Register() {
    const form = new FormData(document.getElementById('register-form'));
    form.get('username-reg')
    form.get('password-reg')
   fetch('/register', {
       method: 'POST',
       body: form
   }).then(function(response) { //https://developer.mozilla.org/en-US/docs/Web/API/Body/text
    return response.text().then(function(text) {
        console.log("text", text);
        elem = document.getElementById('status1');
        elem.innerHTML = text;
    });
   });
}

function Login() {
    const form = new FormData(document.getElementById('login-form'));
    form.get('username-log')
    form.get('password-log')
   fetch('/login', {
       method: 'POST',
       body: form
   }).then(function(response) { //https://developer.mozilla.org/en-US/docs/Web/API/Body/text
    return response.text().then(function(text) {
        if (text === 'true' ) {
            window.location.href = '/';
            return;
        }
        console.log("text", text);
        elem = document.getElementById('status2');
        elem.innerHTML = '<h1> Wrong username or password </h1>';
    });
   });
}
