#!/usr/bin/env python

import json


def find_different_userids(file1, file2):
    stats = {}
    old_stats = {}
    # Read userIds from both files
    with open(file1, "r") as f1, open(file2, "r") as f2:
        for line in f1:
            points, stars, userid = map(int, line.split(":"))
            stats[userid] = (points, stars)
        for line in f2:
            points, stars, userid = map(int, line.split(":"))
            old_stats[userid] = (points, stars)

    with open("leaderboard.json") as json_file:
        leaderboard = json.load(json_file)

    leaders = []
    for usr in stats.keys():
        if stats[usr] != old_stats[usr]:
            leaders.append(
                leaderboard.get("members", {}).get(str(usr), {}).get("name", None)
            )

    output = ""
    for i, leader in enumerate(leaders):
        output += leader
        if i == len(leaders) - 2:
            output += " og "
        elif i < len(leaders) - 1:
            output += ", "

    output += " har klatret på pallen!"
    print(output)


# Example usage:
find_different_userids("stats.old", "stats")
