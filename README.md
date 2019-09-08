# Terminal Slacker
Trying to bring the joy of [irssi](https://irssi.org) to the users of 
[slack](https://slack). No plain and simple just trying to create a new slack
client with a _twist_, It is running in a terminal, and I should try try not
to use all memory in the world.


# Disclaimer
I'm NOT a [go](https://golang.org) programmer in any way, I would like to learn
though. So If you have the time, and is thinking that "Slacking off" in a 
terminal could be great, please join and help me make this perhaps slightly
usable. -- Thanks 

# Installation
```
go get github.com/thorsager/t-slacker
```

# Configuration
Configuration of `t-slacker` is currently very basic, it is done by placing a 
file called `config.json` in the `~/.t-slacker` folder.

The content should look something like this:
```json
{
  "debug": false,
  "notify": false,
  "teams": [
    { "name": "my-team",
      "auto_connect": true,
      "auto_join": true,
      "slack_token": "secret-t0ken",
      "history" : {
        "fetch": true,
	    "size": 24
      },
	  "colorize": true
    },
    { "name": "my-other-team",
      "auto_connect": true,
      "slack_token": "another-secret-t0ken",
      "history" : {
        "fetch": true,
        "size": 24
      },
      "colorize": true
    }
  ]
}
```

## Getting a token.
For now, until something better is implemented authentication is done using 
[Legacy Tokens](https://api.slack.com/custom-integrations/legacy-tokens) 

