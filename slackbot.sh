#!/bin/bash

year=$(date +%Y)
webhook=$(cat slack-webhook.secret)
sessionkey=$(cat session-key.secret)
leaderboard="395034"

function fetch_leaderboard() {
	# create tmp file
	tmpfile=$(mktemp)

	# if successful, move from tmp file to leaderboard.json
	curl -fsS https://adventofcode.com/$year/leaderboard/private/view/$leaderboard.json \
		--cookie "session=$sessionkey" \
		--output $tmpfile && mv $tmpfile leaderboard.json
}

function scoreboard() {
	index=1
	for player in $(cat stats):
	do
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

		((index=$index+1))
	done
}


# fetch latest data
echo "$(date) fetching leaderboard data"
fetch_leaderboard


# create stats file
cat leaderboard.json | jq '.members[] | (.local_score, ":", .stars, ":", .id, "\n")' -j | sort -rn | grep -v '0:0' > stats


# compare with old stats
if cmp -s "stats" "stats.old"; then
	echo "$(date) no changes in leaderboard"
else
	echo "$(date) > detected a change in leaderboard, notifying slack"
	
	curl -fsS $webhook \
		-X POST -H 'Content-type: application/json' \
		--data "{ \"text\": \"Noen har klart en ny oppgave! Ny poengoversikt:\n\n$(scoreboard)\" }" \
		--output /dev/null
fi


# don't forget to save the state
cp stats stats.old
