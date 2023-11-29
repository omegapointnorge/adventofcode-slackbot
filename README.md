# AdventOfCode Slackbot

This is a slackbot that will post a message to our `#advent-of-code` slack
channel when the leaderboard has been updated. It is written as an AWS Lambda
in Go, and runs on a fixed schedule. Previously every 2 hours has worked out
quite nice, so it doesn't spam too often during workdays.


## Work in progress

The program is working locally, but it needs to handle state so that it only
post when there is a change. Also it needs to handle credentials.
