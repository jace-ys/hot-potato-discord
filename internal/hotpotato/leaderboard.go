package hotpotato

import (
	"sort"

	"github.com/jace-ys/hot-potato-discord/internal/room"
)

type UserDeaths struct {
	UserID string
	Count  int
}

type Scoreboard []UserDeaths

func (sb Scoreboard) Len() int           { return len(sb) }
func (sb Scoreboard) Swap(i, j int)      { sb[i], sb[j] = sb[j], sb[i] }
func (sb Scoreboard) Less(i, j int) bool { return sb[i].Count > sb[j].Count }

func BuildLeaderboard(counters []room.DeathCounter, top ...int) Scoreboard {
	leaderboard := make(Scoreboard, len(counters))
	for i, counter := range counters {
		leaderboard[i] = UserDeaths{counter.UserID, counter.Count}
	}

	sort.Sort(leaderboard)

	if len(top) > 0 {
		count := top[0]
		if count <= len(leaderboard) {
			return leaderboard[:count]
		}
	}

	return leaderboard
}
