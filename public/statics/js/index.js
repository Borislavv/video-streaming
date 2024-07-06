const container = document.querySelector('.container');
const listBtn = document.getElementById('list-btn');
const videoList = document.getElementById('video-list');

videoList.style.width = `${videoPlayer.clientWidth}px`;

listBtn.addEventListener('click', function (event) {
    event.stopPropagation();
    videoList.style.display = (videoList.style.display === 'block') ? 'none' : 'block';
    videoList.classList.toggle('active');
    container.classList.toggle('with-list');
});

const paginationInfo = document.querySelector('.pagination-info')
const paginationControl = document.querySelector('.pagination-control')
document.addEventListener('click', function (event) {
    let matched = false
    let ul = document.querySelector('.video-list');
    if (ul !== null) {
        let lis = ul.getElementsByTagName('li');
        Array.from(lis).forEach(function (el) {
            if (event.target === el) {
                matched = true
            }
        });
    }

    if (
        (event.target !== videoList &&
        event.target !== paginationInfo &&
        event.target !== paginationControl &&
        event.target !== limitSelect &&
        event.target !== pageSelect &&
        event.target !== reqBtn) &&
        matched === false
    ) {
        videoList.style.display = 'none';
    }
});

function loadVideoList() {
    let limit = parseInt(limitSelect.value, 10);
    let page  = parseInt(pageSelect.value, 10);

    // clear the previous video list
    const ul = document.querySelector('.dropdown-content ul');
    ul.innerHTML = '';

    let token = getCookie('x-access-token')

    fetch(`http://0.0.0.0:8000/api/v1/video?limit=${limit}&page=${page}`, {
        headers: {
            'x-access-token': token,
        }
    })
    .then(response => response.json())
    .then(data => {
        // render list of videos
        renderList(data.data)
    })
    .catch(error => {
        console.error('Error occurred while loading a video list:', error);
        ul.textContent = 'Sorry, there is an error occurred while loading a video list';
    });
}

function renderList(data) {
    // clear the previous video list
    const ul = document.querySelector('.dropdown-content ul');
    ul.innerHTML = '';

    const paginationInfo = document.querySelector('.dropdown-content .pagination-info');

    // clear the previous pages list
    const pageSelect = document.getElementById('page-select');
    pageSelect.innerHTML = '';

    if (data.list && data.list.length > 0) {
        const videoList = document.createElement('ul');
        videoList.className = 'video-list';

        data.list.forEach(video => {
            const listItem = document.createElement('li');
            listItem.className = 'list-item';
            listItem.textContent = video.name;
            listItem.id = video.id.value
            videoList.appendChild(listItem);
        });

        ul.appendChild(videoList);

        let totalPages = Math.ceil(data.pagination.total / currentLimit);

        paginationInfo.textContent = `Page ${currentPage} of ${totalPages}`;

        // available pages list building
        for (let i = 1; i <= totalPages; i++) {
            const pageListItem = document.createElement('option');
            pageListItem.textContent = `Page ${i}`;
            pageListItem.value = `${i}`;
            pageSelect.appendChild(pageListItem);
        }

    } else {
        ul.textContent = 'There are no available videos';
        ul.style = 'align: center';

        paginationInfo.innerHTML = '';
    }
}

// handling limit 'select' box
const limitSelect = document.getElementById('limit-select');
limitSelect.addEventListener('change', () => {
    let chosenLimit = parseInt(limitSelect.value, 10)

    if (chosenLimit !== currentLimit) {
        previousLimit = currentLimit;
        currentLimit  = chosenLimit;
    }
});

// handling page 'select' box
const pageSelect = document.getElementById('page-select');
pageSelect.addEventListener('change', () => {
    let chosenPage = parseInt(pageSelect.value, 10);

    if (chosenPage !== currentPage) {
        previousPage = currentPage;
        currentPage  = chosenPage;
    }
});

// handling list request btn
const reqBtn = document.getElementById('request-btn');
reqBtn.addEventListener('click', function () {
    if (previousPage !== currentPage || previousLimit !== currentLimit) {
        loadVideoList();
        previousLimit = currentLimit;
        previousPage  = currentPage;
    } else {
        showAlert('There are no changes in page or limit');
    }
});

function getCookie(cookieName) {
    const cookies = document.cookie.split(';');
    for (let i = 0; i < cookies.length; i++) {
        const cookie = cookies[i].trim();
        if (cookie.startsWith(cookieName + '=')) {
            return cookie.substring(cookieName.length + 1);
        }
    }
    return null;
}

/** PopUP */
document.getElementById('video-control').addEventListener('click', function() {
    document.getElementById('video-upload-popup').style.display = 'block';
});

document.querySelector('.close-button').addEventListener('click', function() {
    document.getElementById('video-upload-popup').style.display = 'none';
});

/** PopUp: uploading video to the server */
document.getElementById('video-upload-form').addEventListener('submit', function(event) {
    event.preventDefault();

    let formData = new FormData(this); // 'this' refers to the form
    let xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/v1/resource', true);
    xhr.onload = function (data) {
        if (xhr.status === 200 || xhr.status === 201) {
            console.log('File successfully uploaded.');
            console.log('Response from the server:', JSON.parse(xhr.responseText));
            let response = JSON.parse(xhr.responseText);

            let iData = {
                name: response.data.name,
                resourceID: response.data.id,
                description: "Awesome file :)"
            };

            let ixhr = new XMLHttpRequest();
            ixhr.open('POST', '/api/v1/video', true);
            ixhr.setRequestHeader('Content-Type', 'application/json'); // Установка заголовка для JSON

            ixhr.onload = function () {
                if (ixhr.status === 200 || ixhr.status === 201) {
                    console.log('Video successfully created.');
                    console.log('Response from the server:', JSON.parse(ixhr.responseText));
                } else {
                    console.error('An error occurred during the video creation.');
                }
            };

            ixhr.onerror = function () {
                console.error('An error occurred during the request.');
            };

            ixhr.send(JSON.stringify(iData)); // Отправка данных в формате JSON
        } else {
            console.error('An error occurred during the file uploading.');
        }
    };
    xhr.send(formData);


});

// init. default data
let currentLimit    = parseInt(limitSelect.value, 10);
let currentPage     = parseInt(pageSelect.value, 10);
let previousLimit   = currentLimit;
let previousPage    = currentPage;
loadVideoList();