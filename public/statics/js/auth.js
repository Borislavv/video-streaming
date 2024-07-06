document.getElementById('main-form').addEventListener('submit', function (event) {
    event.preventDefault();

    const formTitle = document.getElementById('form-title').textContent;
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;

    const userDetails = {
        email: email,
        password: password
    };

    if (formTitle === "Registration") {
        userDetails.username = document.getElementById('username').value;
        userDetails.birthday = document.getElementById('birthday').value;
    }

    const endpoint = formTitle === "Login" ? "/api/v1/authorization" : "/api/v1/registration";

    fetch(`http://0.0.0.0:8000${endpoint}`, {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(userDetails)
    })
        .then(response => response.json())
        .then(handleResponse)
        .catch(error => {
            console.error('Error occurred while loading a video list:', error);
            showErrorMessage('Error: ' + error.message);
            setTimeout(clearErrorMessage, 5000);
        });
});

function switchToLogin() {
    document.getElementById('form-title').textContent = "Login";
    document.getElementById('extra-fields').innerHTML = "";
}

function switchToRegistration() {
    document.getElementById('form-title').textContent = "Registration";
    document.getElementById('extra-fields').innerHTML = `
        <label for="username">Username:</label>
        <input type="text" id="username" name="username" required>

        <label for="birthday">Birthday:</label>
        <input type="date" id="birthday" name="birthday" required>
    `;
}

function showErrorMessage(message) {
    const errorText = document.getElementById('error-text');
    const errorMessage = document.getElementById('error-message');
    const errorClose = document.getElementById('error-close');
    errorText.textContent = message;
    errorText.style.display = 'block';
    errorMessage.style.display = 'block';
    errorClose.style.display = 'block';
}

function clearErrorMessage() {
    const errorText = document.getElementById('error-text');
    const errorMessage = document.getElementById('error-message');
    const errorClose = document.getElementById('error-close');
    errorText.style.display = 'none';
    errorMessage.style.display = 'none';
    errorClose.style.display = 'none';
}

function handleResponse(data) {
    if (data.error) {
        showErrorMessage(data.error.message || "A server error occurred");
        setTimeout(clearErrorMessage, 5000); // Автоматическое скрытие сообщения об ошибке через 5 секунд
    } else {
        document.cookie = `x-access-token=`+data.data;
        window.location.replace("/");
    }
}

document.getElementById('close-error').onclick = () => {
    document.getElementById('error-message').style.display = 'none';
};
