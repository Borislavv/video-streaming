const websocket = new WebSocket('ws://0.0.0.0:9988/');

const videoPlayer = document.getElementById('videoPlayer');
const nextBtn = document.getElementById('next-btn');
const prevBtn = document.getElementById('prev-btn');

websocket.binaryType = 'arraybuffer';

let buffer;
let mediaSource;
let chunks;
let mediaSourceReady;
let token;

// ws event: open
websocket.onopen = (event) => {
    console.log('WebSocket connection opened');

    token = getCookie('x-access-token')
    if (token === "") {
        throw "token is not provided";
    }
};
// ws event: close
websocket.onclose = (event) => {
    console.log('WebSocket connection closed');
};
// ws event: error
websocket.onerror = (event) => {
    console.error('WebSocket error: ', event);
};

// ws event: on message
websocket.onmessage = (event) => {
    const data = event.data;

    console.log("Some data received...", data)

    if (typeof data === 'string' && (
        data.startsWith('start') ||
        data.startsWith('error') ||
        data.startsWith('stop')
    )) {
        console.log('Data is action: ' + data)

        if (data.startsWith('start')) {
            console.log("Starting new video...")
            let dataParts = data.split('::')
            console.log(dataParts)
            makeMediaResource(dataParts[1], dataParts[2])
        } else if (data.startsWith('error')) {
            let dataParts = data.split('::')
            console.log("Server error occurred: " + dataParts[1])
            showAlert(dataParts[1])
        } else if (data === 'stop') {
            console.log("Stopping playing...")
            closeMediaResource()
        }

        return;
    }

    if (data instanceof ArrayBuffer) {
        console.log('Data is chunk, adding to buffer...')
        chunks.push(data)
        addNextChunk()
    }
};

videoPlayer.addEventListener('seeking', function (event) {
    console.log("---> REQUEST FROM: ", event.currentTarget.currentTime, event.currentTarget.duration)

    // socket.send("ID:"+currentVideoID+":FROM:"+event.currentTarget.currentTime+":TO:"+event.currentTarget.duration)
});

// todo need to make a ticker which will count the awaiting time if it's more than N then send decrease buffer action
videoPlayer.addEventListener('waiting', function() {
    console.warn('Video playback is waiting for data (buffering)');

    // requesting decrease the buffer
    // socket.send("decrBuff") // not implemented yet
});

let currentVideoID = ''
// initialization function
function waitForVideoListWillBeRendered(selector, callback) {
    const element = document.querySelector(selector);

    if (element) {
        callback(element);
    } else {
        const checkElement = () => {
            const element = document.querySelector(selector);
            if (element) {
                callback(element);
            } else {
                requestAnimationFrame(checkElement);
            }
        };
        requestAnimationFrame(checkElement);
    }
}

// called init. function
waitForVideoListWillBeRendered('.video-list', function () {
    let UL = document.querySelector('.video-list');
    if (UL !== null) {
        let LIs = UL.getElementsByTagName('li');
        let isSettingUp = false;
        Array.from(LIs).forEach(function (li) {
            if (!isSettingUp) {
                currentVideoID = li.id;
                isSettingUp = true;
                console.log('Requesting the first video ' + "ID:" + currentVideoID)
                console.log(li)
                requestByID('ID', currentVideoID)
            }
        });
    }
});

// next video handler
nextBtn.addEventListener('click', function() {
    let UL = document.querySelector('.video-list');
    let found = false
    if (UL !== null) {
        let LIs = UL.getElementsByTagName('li');
        Array.from(LIs).forEach(function (li) {
            if (found) {
                console.log('Requesting the next video ' + "ID:" + li.id)
                requestByID('ID', li.id)
                currentVideoID = li.id // updating the current video ID
                found = false // skipping further iterations
            } else {
                if (currentVideoID === '') { // setting up the current ID if it wasn't set up
                    currentVideoID = li.id
                }
                if (currentVideoID === li.id) { // setting up the flag which helps to determine
                    found = true                // that current video was found, and we need to take the next ID
                }
            }
        });
    }
});

// previous video handler
prevBtn.addEventListener('click', function(event) {
    let UL = document.querySelector('.video-list');
    if (UL !== null) {
        let LIs= UL.getElementsByTagName('li');
        let found           = false
        let previousVideoID = ''
        Array.from(LIs).forEach(function (li) {
            if (found) {
                return; // skipping unnecessary iterations
            }
            if (currentVideoID === '') { // setting up the current ID if it's empty
                currentVideoID = li.id
            }
            if (previousVideoID === '') { // setting up the previous ID if it's empty
                previousVideoID = li.id
            }
            if (currentVideoID === li.id) { // checking up that we found the target video
                currentVideoID = previousVideoID // updating the actual video ID
                console.log('Requesting the previous video ' + "ID:" + currentVideoID)
                // requesting the prev video/audio from server
                requestByID('ID', currentVideoID)
                found = true // setting up the var. for skipp unnecessary iterations
            } else {
                previousVideoID = li.id // updating the previous video ID
            }
        });
    }
});

// by id video handler
document.addEventListener('click', function (event) {
    let UL = document.querySelector('.video-list');
    if (UL !== null) {
        let LIs = UL.getElementsByTagName('li');
        Array.from(LIs).forEach(function (li) {
            if (currentVideoID === '') { // setting up the current video ID if it wasn't set up
                currentVideoID = li.id
            }
            if (event.target === li && currentVideoID !== li.id) { // check the target video ID is not equals with the current
                // requesting the prev video/audio from server
                requestByID('ID', li.id)
                currentVideoID = li.id; // updating the current video ID
            }
        });
    }
});

function requestByID(strategy, id) {
    let data = `${strategy}::{ "id": "${id}", "token": "${token}" }`
    console.log("websocket request: " + data);
    console.log(id)
    websocket.send(data)
}

function addNextChunk() {
    awaiting()
        .then(
            function () {
                try {
                    buffer.appendBuffer(chunks.shift())
                    console.log("Chunk successfully added to buffer")
                } catch (error) {
                    console.error("Chunk adding to buffer failed", error)
                }
            }
        ).catch(
            function (e) {
                console.error("unable to awaiting", e)
            }
        )
}

async function awaiting() {
    while(!mediaSourceReady || buffer.updating || chunks.length === 0) {
        console.log(
            "Awaiting data: " +
            "media resource is not ready yet -", mediaSourceReady, ", " +
            "buffer is still updating -", buffer.updating,", " +
            "chunks awaiting -", chunks.length === 0,
            "(len:", chunks.length, ")"
        )
        await new Promise(r => setTimeout(r, 1));
    }
}

function closeMediaResource() {
    awaitingClose()
        .then(
            function () {
                try {
                    mediaSource.endOfStream()
                    console.log("Stream is successfully closed")
                } catch (error) {
                    console.error("Stream closing failed", error, mediaSource.readyState)
                }
            }
        ).catch(
            function (e) {
                console.error("unable to awaiting closing", e)
            }
        )
}

async function awaitingClose() {
    while(buffer.updating || chunks.length > 0) {
        console.log(
            "Awaiting closing: " +
            "buffer is still updating -", buffer.updating,", " +
            "chunks is not empty -", chunks.length > 0,
                "(len:", chunks.length, ")"
        )
        await new Promise(r => setTimeout(r, 1));
    }
}

function makeMediaResource(audioCodec, videoCodec) {
    mediaSource = new MediaSource();
    mediaSourceReady = false;
    videoPlayer.src = URL.createObjectURL(mediaSource);
    chunks = [];

    mediaSource.addEventListener('sourceclose', function (e) {
        console.log("MediaSource sourceopen event is closed", e)
    });

    mediaSource.addEventListener('sourceopen', function () {
        console.log("MediaSource sourceopen event is open");

        try {
            let codecsMap = []
            codecsMap['avc1'] = 'avc1.42E01E'
            codecsMap['mp4a'] = 'mp4a.40.2'

            let codecsStr = ''
            if (videoCodec !== '' && codecsMap[videoCodec]) {
                codecsStr += codecsMap[videoCodec]
            }
            if (audioCodec !== '' && codecsMap[audioCodec]) {
                if (videoCodec !== '') {
                    codecsStr += ', '
                }
                codecsStr += codecsMap[audioCodec]
            }
            if (codecsStr === '') {
                console.error('Codecs string a empty! Unable to play video!')
            }

            const codec = 'video/mp4; codecs="' + codecsStr +'"';

            console.log("CODEC: ", codec)

            if (MediaSource.isTypeSupported(codec)) {
                console.log("Type of codec is supported: ", codec);

                if (mediaSource.sourceBuffers.length > 0) {
                    let b = mediaSource.sourceBuffers.length
                    while(b > 1) {
                        mediaSource.removeSourceBuffer(mediaSource.sourceBuffers[b]);
                        b--
                    }
                }

                buffer = mediaSource.addSourceBuffer(codec);
                console.log("Source buffer added", buffer);
            } else {
                console.log("Type of codec is NOT supported: ", codec);
            }
        } catch (e) {
            console.error("Error adding source buffer:", e);
        }

        buffer.addEventListener('error', function (e) {
            console.error('Buffer error', e)
        });

        mediaSourceReady = true;
    }, false);
}

function showAlert(message) {
    const alertContainer = document.getElementById('custom-alert');
    const alertMessage   = document.getElementById('alert-message');
    const closeButton    = document.getElementById('close-alert');

    alertMessage.textContent     = message;
    alertContainer.style.display = 'flex';

    closeButton.addEventListener('click', () => {
        alertContainer.style.display = 'none';
    });

    setTimeout(() => {
        alertContainer.style.display = 'none';
    }, 5000);
}