package game

const (
	POKNINE       int = 1
	POKEIGHT      int = 2
	TONG          int = 3
	STRAIGHTFLUSH int = 4
	STRAIGHT      int = 5
	SIAN          int = 6
	POINT         int = 7
)

// A PlayerState table. Active means is this the player's turn.
var PokdengPlayerState = struct {
	Playing uint8 // Start playing but not active.
	WaitPok uint8
	EndGame uint8
}{1, 2, 3}

type PokdengPlayer struct {
	UID                 string
	HandCardList        CardList // On hand CardList. (Negative Point)
	playerState         uint8    // Player's state. Will be NotPlaying, NotActivePlaying, ActiveWaitBet, ActiveWaitPok, ActiveWaitJua
	PlayerNeedToStandUp bool
	Bet                 uint64
	ChipGain            int64
	SeatPosition        uint8
	IsJub               bool
}

func NewPokdengPlayer() *PokdengPlayer {
	return &PokdengPlayer{
		UID:                 "",
		HandCardList:        CardList{},
		playerState:         PokdengPlayerState.Playing,
		PlayerNeedToStandUp: false,
		Bet:                 0,
		ChipGain:            0,
		SeatPosition:        0,
	}
}

func (p *PokdengPlayer) cleanup() {
	p.HandCardList = CardList{}
}

func (p *PokdengPlayer) IsPok() bool {
	return p.HandCardList.IsPok()
}

func (p *PokdengPlayer) GetCardPoint() int8 {
	return p.HandCardList.GetPokdengCardPoint()
}

func (p *PokdengPlayer) GetCardResultAndDeng() (int, int, int8) {
	pointMultiply := 1
	if p.IsPok() {
		if p.GetCardPoint() == 9 {
			sameSuit, _ := p.HandCardList.IsSameSuit()
			if sameSuit {
				pointMultiply = 2
			}
			return POKNINE, pointMultiply, 0
		}
		sameSuit, _ := p.HandCardList.IsSameSuit()
		if sameSuit {
			pointMultiply = 2
		}

		sameNumber, _ := p.HandCardList.IsSameNumber()
		if sameNumber {
			pointMultiply = 2
		}
		return POKEIGHT, pointMultiply, 0

	} else if p.HandCardList.IsThreeOfAKind() {
		pointMultiply = 5
		return TONG, pointMultiply, 0
	} else if p.HandCardList.IsSequence() {
		pointMultiply = 3

		sameSuit, _ := p.HandCardList.IsSameSuit()
		if sameSuit {
			pointMultiply = 5
			return STRAIGHTFLUSH, pointMultiply, 0
		}
		return STRAIGHT, pointMultiply, 0
	} else if p.HandCardList.IsSian() {
		pointMultiply = 3
		return SIAN, pointMultiply, 0
	}

	sameSuit, _ := p.HandCardList.IsSameSuit()
	sameNumber, _ := p.HandCardList.IsSameNumber()

	if sameSuit {
		if len(p.HandCardList.cards) == 2 {
			pointMultiply = 2
		} else if len(p.HandCardList.cards) == 3 {
			pointMultiply = 3
		}
	} else if sameNumber {
		pointMultiply = 2
	}
	return POINT, pointMultiply, p.GetCardPoint()
}

func (p *PokdengPlayer) UpdateStatistic() {

}

func (p *PokdengPlayer) GetPlayerState() uint8 {
	return p.playerState
}
