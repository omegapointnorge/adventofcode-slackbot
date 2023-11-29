package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"
)

const (
	// TODO: Handle secres with a cloud credentials manager
	leaderboardUrl = "https://adventofcode.com/2023/leaderboard/private/view/395034.json"
	// TODO: Change this to use correct channel. See https://api.slack.com/apps/A04DSQ8FPLY/incoming-webhooks
	secretSlackWebhook = ""
	// TODO: Double check this
	secretSessionKey = ""
)

type (
	LeaderboardDTO struct {
		Members map[string]Member
	}

	Leaderboard struct {
		Members []Member
	}

	Member struct {
		Id    int
		Name  string
		Score int `json:"local_score"`
		Stars int
	}

	SlackBot struct {
		HTTPClient HTTPClient
	}

	HTTPClient interface {
		Do(req *http.Request) (*http.Response, error)
	}

	SlackWebhook struct {
		Text string `json:"text"`
	}
)

func NewSlackBot(httpClient HTTPClient) *SlackBot {
	return &SlackBot{HTTPClient: httpClient}
}

func (lb *Leaderboard) SortByHighestScoreAndStars() {
	sort.Slice(lb.Members, func(i, j int) bool {
		m1 := lb.Members[i]
		m2 := lb.Members[j]
		return m1.Score+m1.Stars > m2.Score+m2.Stars
	})
}

func (lb *Leaderboard) IsEqualTo(lb2 *Leaderboard) bool {
	if len(lb.Members) != len(lb2.Members) {
		return false
	}
	for i, m1 := range lb.Members {
		if m1.Score != lb2.Members[i].Score {
			return false
		}
		if m1.Name != lb2.Members[i].Name {
			return false
		}
	}
	return true
}

func (bot *SlackBot) GetLeaderboard() (*Leaderboard, error) {
	jsonData, err := bot.GetLeaderboardJson()
	if err != nil {
		return nil, err
	}

	var lbDto LeaderboardDTO
	err = json.Unmarshal([]byte(jsonData), &lbDto)
	if err != nil {
		return nil, err
	}

	// Copy values from map into a slice
	var lb Leaderboard
	for _, member := range lbDto.Members {
		lb.Members = append(lb.Members, member)
	}

	lb.SortByHighestScoreAndStars()

	return &lb, nil
}

func (bot *SlackBot) GetLeaderboardJson() (string, error) {
	req, err := http.NewRequest("GET", leaderboardUrl, nil)
	if err != nil {
		return "", err
	}

	req.AddCookie(&http.Cookie{Name: "session", Value: secretSessionKey})

	resp, err := bot.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP request failed with status code %d", resp.StatusCode)
	}

	return string(body), nil
}

func (m *Member) FormatText() string {
	var name string
	if m.Name == "" {
		name = fmt.Sprintf("Anonym bruker (%d)", m.Id)
	} else {
		name = m.Name
	}
	return fmt.Sprintf("*%s* Poeng: %d, :star: %d", name, m.Score, m.Stars)
}

func (lb *Leaderboard) FormatText() string {
	var text string

	for i, member := range lb.Members {
		if i > 0 {
			text += "\n"
		}

		emoji := ""
		if member.Score > 0 {
			switch i {
			case 0:
				emoji = ":first_place_medal: "
			case 1:
				emoji = ":second_place_medal: "
			case 2:
				emoji = ":third_place_medal: "
			default:
				emoji = fmt.Sprintf(":number-%d: ", i+1)
			}
		}

		text += fmt.Sprintf("%s%s", emoji, member.FormatText())
	}

	return text
}

func main() {
	client := &http.Client{Timeout: time.Second * 10}
	slackBot := NewSlackBot(client)

	// Call AdventOfCode
	lb, err := slackBot.GetLeaderboard()
	if err != nil {
		panic(err)
	}

	// TODO: Check if leaderboard has changed, and if not, exit main function here

	// Create Slack payload
	msg := SlackWebhook{
		Text: "Noen har klart en ny oppgave! Ny poengoversikt:\n" + lb.FormatText(),
	}

	// Convert to JSON
	body, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	// Trigger the Slack webhook
	response, err := client.Post(secretSlackWebhook, "application/json", bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	// Check for errors
	if response.StatusCode != 200 {
		fmt.Printf("Slack returned an HTTP %s error\n", response.Status)
		return
	}

	fmt.Println("Successfully posted the leaderboard to slack")

	// TODO: Update leaderboard state
}
