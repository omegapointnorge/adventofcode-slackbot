#!/bin/bash

year=$(date +%Y)
webhook=$(cat slack-webhook.secret)
sessionkey=$(cat session-key.secret)
gptApiKey=$(cat gpt-api-key.secret)
leaderboard="395034"

function fetch_leaderboard() {
	# create tmp file
	tmpfile=$(mktemp)

	# if successful, move from tmp file to leaderboard.json
	curl -fsS https://adventofcode.com/$year/leaderboard/private/view/$leaderboard.json \
		--cookie "session=$sessionkey" \
		--output $tmpfile && mv $tmpfile leaderboard.json
}

function generate_commentary() {
	gptEndpoint="https://api.openai.com/v1/chat/completions"

	# Create a JSON request payload with the necessary prompts
	gptRequestPayload=$(jq -n --arg stats_old "$(cat stats.old)" --arg stats_new "$(cat stats)" --arg leaderboard_json "$(cat leaderboard.json)" '{
        "model": "gpt-4",
        "messages": [
            {
                "role": "system",
                "content": "We are participating in this years Advent Of Code at work. You are a slackbot for our work channel. Provide short commentary on leaderboard changes. Rate by local_score then stars. Dont post rankings. Avoid inappropriate content and errors. Dont post user IDs"
            },
            {
                "role": "user",
                "content": "Here are the old stats, new stats and current leaderboard:\n\($stats_old)\n\n\($stats_new)\n\n\($leaderboard_json)"
            }
        ]
    }')

	# Send the request to the API
	apiResponse=$(curl --location --request POST $gptEndpoint \
		--header "Authorization: Bearer $gptApiKey" \
		--header "Content-Type: application/json" \
		--data-raw "$gptRequestPayload")

	# Extract content using jq
	commentary=$(echo "$apiResponse" | jq -r '.choices[0].message.content')

	# Print the extracted commentary
	echo "$commentary"
}

function scoreboard() {
	index=1
	for player in $(cat stats); do
		# get points, stars and player id
		points=$(echo $player | cut -d: -f1)
		stars=$(echo $player | cut -d: -f2)
		id=$(echo $player | cut -d: -f3)

		# get the name
		name=$(cat leaderboard.json | jq -r ".members[] | select(.id==$id) | .name")

		# anonymous user?
		if [ "$name" == "null" ]; then
			name="Anonym bruker ($id)"
		fi

		# top 3 gets an emoji medal
		case $index in
		1)
			emoji=":first_place_medal:"
			;;
		2)
			emoji=":second_place_medal:"
			;;
		3)
			emoji=":third_place_medal:"
			;;
		**)
			emoji=":number-${index}:"
			;;
		esac

		# ..but only when there are some points earned
		if [ "$points" -eq 0 ]; then
			emoji=""
		fi

		# prettify
		echo "${emoji} *$name* Poeng: $points, :star: $stars\n"

		((index = $index + 1))
	done
}

# fetch latest data
echo "$(date) fetching leaderboard data"
fetch_leaderboard

# create stats file
cat leaderboard.json | jq '.members[] | (.local_score, ":", .stars, ":", .id, "\n")' -j | sort -rn | grep -v '0:0' >stats

# compare with old stats
if cmp -s "stats" "stats.old"; then
	echo "$(date) no changes in leaderboard"
else
	echo "$(date) > detected a change in leaderboard, notifying slack"

	# Fetch commentary
	commentary=$(generate_commentary)

	# Check if commentary is not empty and not null
	if [ -z "$commentary" ] || [ "$commentary" == "null" ]; then
		echo "Commentary generation failed or returned null. Setting commentary to an empty string."
		commentary=""
	fi

	text=$(python3 compare_users.py)

	# Post update to slack
	curl -fsS $webhook \
		-X POST -H 'Content-type: application/json' \
		--data "{ \"text\": \"$text\", \"attachments\": [{\"text\": \"Ny poengoversikt:\n\n$(scoreboard)\n\n:robot_face: ChatGPT:\n\n$commentary\"}] }" \
		--output /dev/null
fi

# Don't forget to save the state
cp stats stats.old
