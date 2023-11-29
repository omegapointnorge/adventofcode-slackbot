package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDetectSwappedMembers(t *testing.T) {
	old := &Leaderboard{[]Member{{Name: "alice", Score: 0}}}
	new := &Leaderboard{[]Member{{Name: "arnold", Score: 0}}}
	assertFalse(t, new.IsEqualTo(old))
}

func TestDetectNewMember(t *testing.T) {
	old := &Leaderboard{[]Member{{Name: "alice", Score: 0}}}
	new := &Leaderboard{[]Member{{Name: "alice", Score: 0}, {Name: "bob", Score: 0}}}
	assertFalse(t, new.IsEqualTo(old))
}

func TestDetectIncreasedScore(t *testing.T) {
	old := &Leaderboard{[]Member{{Name: "alice", Score: 0}}}
	new := &Leaderboard{[]Member{{Name: "alice", Score: 10}}}
	assertFalse(t, new.IsEqualTo(old))
}

func TestNoChange(t *testing.T) {
	old := &Leaderboard{[]Member{{Name: "alice", Score: 0}}}
	new := &Leaderboard{[]Member{{Name: "alice", Score: 0}}}
	assertTrue(t, new.IsEqualTo(old))
}

func TestFormatTextMemberWithZeroScore(t *testing.T) {
	m := Member{Name: "alice", Score: 0, Stars: 0}
	assertEquals(t, "*alice* Poeng: 0, :star: 0", m.FormatText())
}

func TestNoEmojisWhenAllZeros(t *testing.T) {
	lb := &Leaderboard{[]Member{{Name: "alice"}, {Name: "bob"}, {Name: "charlie"}}}

	text := lb.FormatText()

	expected := `*alice* Poeng: 0, :star: 0
*bob* Poeng: 0, :star: 0
*charlie* Poeng: 0, :star: 0`

	assertEquals(t, expected, text)
}

func TestMedalEmojisForTopThree(t *testing.T) {
	lb := &Leaderboard{
		[]Member{
			{Name: "alice", Score: 50, Stars: 5},
			{Name: "bob", Score: 40, Stars: 4},
			{Name: "charlie", Score: 30, Stars: 3},
			{Name: "david", Score: 20, Stars: 2},
		},
	}

	text := lb.FormatText()

	expected := `:first_place_medal: *alice* Poeng: 50, :star: 5
:second_place_medal: *bob* Poeng: 40, :star: 4
:third_place_medal: *charlie* Poeng: 30, :star: 3
:number-4: *david* Poeng: 20, :star: 2`

	assertEquals(t, expected, text)
}

type MockHTTPClient struct{}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	// Create a mock response
	mockResponse := httptest.NewRecorder()
	mockResponse.WriteHeader(http.StatusOK)

	jsonData := `
  {
    "members": {
      "395": {
        "global_score": 0,
        "completion_day_level": {},
        "id": 395,
        "last_star_ts": 0,
        "local_score": 50,
        "stars": 5,
        "name": "alice"
      },
      "236": {
        "local_score": 50,
        "completion_day_level": {},
        "id": 236,
        "last_star_ts": 0,
        "global_score": 0,
        "name": "bob",
        "stars": 6
      },
      "198": {
        "local_score": 20,
        "completion_day_level": {},
        "id": 198,
        "last_star_ts": 0,
        "global_score": 0,
        "name": "charlie",
        "stars": 2
      },
      "325": {
        "local_score": 20,
        "completion_day_level": {},
        "id": 325,
        "last_star_ts": 0,
        "global_score": 0,
        "name": "david",
        "stars": 3
      },
      "523": {
        "local_score": 1,
        "completion_day_level": {},
        "id": 523,
        "last_star_ts": 0,
        "global_score": 0,
        "name": null,
        "stars": 0
      }
    },
    "owner_id": 395,
    "event": "2023"
	}`

	// Write the JSON payload to the response
	mockResponse.Body.Write([]byte(jsonData))

	// Return the mock response
	return mockResponse.Result(), nil
}

func TestGetLeaderboardShouldSortMembersByHighestScoreAndStars(t *testing.T) {
	mockClient := &MockHTTPClient{}
	slackBot := NewSlackBot(mockClient)
	lb, _ := slackBot.GetLeaderboard()
	assertEquals(t, "bob", lb.Members[0].Name)
	assertEquals(t, "alice", lb.Members[1].Name)
	assertEquals(t, "david", lb.Members[2].Name)
	assertEquals(t, "charlie", lb.Members[3].Name)
}

func TestAnonymousUser(t *testing.T) {
	mockClient := &MockHTTPClient{}
	slackBot := NewSlackBot(mockClient)
	lb, _ := slackBot.GetLeaderboard()
	assertEquals(t, "*Anonym bruker (523)* Poeng: 1, :star: 0", lb.Members[4].FormatText())
}

func assertEquals(t *testing.T, expected string, actual string) {
	if expected != actual {
		t.Errorf("Expected %s, but got %s\n", expected, actual)
	}
}

func assertTrue(t *testing.T, actual bool) {
	if !actual {
		t.Error("Expected true, but was false\n")
	}
}

func assertFalse(t *testing.T, actual bool) {
	if actual {
		t.Error("Expected false, but was true\n")
	}
}
