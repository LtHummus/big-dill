function openVote() {
    fetch('/vote_open', {
        method: 'POST',
    }).then((res) => {
        console.log(res);
    })
}

function closeVote() {
    fetch('/vote_close', {
        method: 'POST',
    }).then((res) => {
        console.log(res);
    })
}