# AdventOfCode Slackbot

This is a script checks a leaderboard every 15 minutes and if there is a change since last time it posts the scores to a slack channel.


## Secrets

The script needs two secrets, which will be read from a file.

1. `slack-webhook.secret` should contain the incoming webhook url for slack.
2. `session-key.secret` should contain a key from the adventofcode session cookie.


## Cronjob

The following cron expression will run the script every 15 minutes every day in December until the 25th:

```
*/15 * 1-25 12 * /root/slackbot.sh >> /root/slackbot.log
```

Use the `crontab -e` command to add it to your server.
