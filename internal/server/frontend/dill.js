async function getWebsocketURL() {
    const resp = await fetch('/socket_url');
    const respJSON = await resp.json();

    return respJSON.socket_url;
}

function makeVoteKey(length) {
    let result           = '';
    const characters       = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    const charactersLength = characters.length;
    for (let i = 0; i < length; i++) {
        result += characters.charAt(Math.floor(Math.random() * charactersLength));
    }
    return result;
}

let clientID;
let voteKey;
let socket;
let votePanel;
let stickerPanel;
let closedPanel;

function showVoteClosedPanel() {
    closedPanel.style.display = 'block';
}

function hideVoteClosedPanel() {
    closedPanel.style.display = 'none';
}

function hideVotePanel() {
    votePanel.style.display = 'none';
}

function showVotePanel() {
    votePanel.style.display = 'block';
}

function showIVotedSticker() {
    stickerPanel.style.display = 'block';
}

function hideIVotedSticker() {
    stickerPanel.style.display = 'none';
}

function getOrSetVoteKey() {
    const key = window.sessionStorage.getItem('vote_key');
    if (key) {
        console.log(`using vote key ${key}`)
        voteKey = key
        return
    }

    console.log('generating vote key');

    voteKey = makeVoteKey(32);
    console.log(`generated ${voteKey}`);
    window.sessionStorage.setItem('vote_key', voteKey);
    return voteKey
}


function messageHandler(evt) {
    console.log(`message recv'd: ${evt.data}`);

    const messagePayload = JSON.parse(evt.data);
    // lol
    const messageKind = messagePayload.kind;

    if (messageKind === 'connect_success') {
        clientID = messagePayload.payload.client_id;
        console.log(`set client id to ${clientID}`)
        document.getElementById('connect_status').style.display = 'block';
    } else if (messageKind === 'vote_status_change') {
        const votesOpen = messagePayload.payload.new_status;
        if (votesOpen) {
            showVotePanel();
            hideIVotedSticker();
            hideVoteClosedPanel();
        } else {
            hideVotePanel();
            showVoteClosedPanel();
            hideIVotedSticker();
        }
    } else if (messageKind === 'vote_success') {
        showIVotedSticker();
    } else if (messageKind === 'vote_status') {
        const status = messagePayload.payload.status;
        if (status === 'votes_closed') {
            showVoteClosedPanel()
        } else if (status === 'already_voted') {
            showIVotedSticker()
        } else {
            showVotePanel()
        }
    } else {
        console.log('unknown message')
    }
}

function connectToWebsocket(url) {
    socket = new WebSocket(url);
    socket.addEventListener('open', (evt) => {
        console.log('connected');
        socket.send(JSON.stringify({
            kind: 'query_vote_status',
            payload: {
                'vote_key': voteKey
            }
        }))
    })

    socket.addEventListener('message', messageHandler)
}

function submitVote(vote) {
    console.log(`will submit vote ${vote}`)
    socket.send(JSON.stringify({
        kind: 'vote',
        payload: {
            vote: vote,
            vote_key: voteKey
        }
    }))
    hideVotePanel();
}

window.onload = async function() {
    console.log('hello world');
    votePanel = document.getElementById('vote_panel')
    stickerPanel = document.getElementById('i_voted_panel')
    closedPanel = document.getElementById('vote_closed_panel')
    getOrSetVoteKey();
    const socketURL = await getWebsocketURL();
    console.log(socketURL);
    connectToWebsocket(socketURL);
}