const passwordField = document.getElementById('pwd-input')

function openVote() {
    const pwd = passwordField.value;
    fetch('/vote_open', {
        method: 'POST',
        headers: {
            'X-Token': pwd,
        }
    }).then((res) => {
        console.log(res);
    })
}

function closeVote() {
    const pwd = passwordField.value;
    fetch('/vote_close', {
        method: 'POST',
        headers: {
            'X-Token': pwd,
        }
    }).then((res) => {
        console.log(res);
    })
}