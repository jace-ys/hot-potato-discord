package hotpotato

import (
	"fmt"
	"math"
	"math/rand"
)

type Potato interface {
	fmt.Stringer
	Kind() string
	PercentChance() int
}

func (gm *GameMaster) GetPotato(kind string) (Potato, error) {
	for _, potato := range gm.potatoes {
		if potato.Kind() == kind {
			return potato, nil
		}
	}

	return nil, ErrInvalidPotatoKind
}

func (gm *GameMaster) RandomPotato() Potato {
	i := rand.Intn(len(gm.potatoes))
	return gm.potatoes[i]
}

func (gm *GameMaster) DecideExplode(potato Potato, turn, heatLevel int) bool {
	inc := math.Log10(math.Pow(float64(heatLevel), float64(turn)))
	chance := potato.PercentChance() + int(inc)
	return rand.Intn(100) <= chance
}

type RawPotato struct{}

func (p RawPotato) String() string {
	return "raw potato"
}

func (p RawPotato) Kind() string {
	return "raw"
}

func (p RawPotato) PercentChance() int {
	return 1
}

type BakedPotato struct{}

func (p BakedPotato) String() string {
	return "baked 🔥 potato 🥔"
}

func (p BakedPotato) Kind() string {
	return "baked"
}

func (p BakedPotato) PercentChance() int {
	return 2
}

type HotPotato struct{}

func (p HotPotato) String() string {
	return "hot 🔥🔥 potato 🥔"
}

func (p HotPotato) Kind() string {
	return "hot"
}

func (p HotPotato) PercentChance() int {
	return 5
}

type BurntPotato struct{}

func (p BurntPotato) String() string {
	return "burnt 🔥🔥🔥 potato 🥔"
}

func (p BurntPotato) Kind() string {
	return "burnt"
}

func (p BurntPotato) PercentChance() int {
	return 10
}
