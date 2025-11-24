# Big Dill

One of the side things I do is help out with the PAX Omegathon. The tl;dr is that PAX (Penny Arcade Expo) is a weekend long
conventions celebrating all things games -- video games, board games, and tabletop games like D&D. One of the features of the
weekend is a tournament called the Omegathon. A handful of folks that bought tickets are randomly selected to compete in a tounrment
throughout the weekend in a variety of games.

I help run the Omegathon as an assistent. For PAX Unplugged 2025, one of the rounds was the game [I'm Kind of A Big Dill](https://prolificgames.net/products/im-kind-of-a-big-dill).
The way this game works is one player at a time is given a prompt to describe certain traits about them such as 
"I am good at home repair" or "I like going to nightclubs". They are also randomly given an "exaggeration token" which contains
a number -3 to +3. A zero tells the player to be completely honest. A positive number tells the player to exaggerate their abilities
and the degree for exaggeration. A negative number has the speaker undersell the abilities. 

The other players in the game listen and then have to guess what number the player has been given and how much they are
exaggerating themselves. 

For the Omegathon, since the audience is watching, the person who runs it had the idea to include some audience voting on
their phones. If the speaker could get a plurality of the audience to guess the correct number, some bonus points would be inorder.

Originally, the plan was to use something like Kahoot, but it only supports a max of 6 answers (we needed 7). There are other ones
but they can only do one question at a time and we weren't sure how many questions we'd need because of potential tie breakers.

I found out about this project on Wednesday before the convention. The round was on Friday. I had my flight on Thuresday....

This is the result of programming my little heart out on the plane on the way to the convention. 

And Friday night rolled around for the _I'm Kind of A Big Dill_ round ... and it worked _perfectly_.

Note that this isn't the greatest code -- the admin panel for opening and closing the vote is pretty barebones and the results
page doesn't live update. User state is kept on device (really just a device ID that we can use to restore state). The frontend
shows and hides elements via CSS and plain 'ol Javascript. Possible choices are hardcoded. etc etc. This code was written 
for exactly one purpose for that Omegathon round and it worked great and I rode that high for the rest of the weekend :)

(Oh, and since I know someone will ask, there was no AI assistance for this project)