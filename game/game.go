package game

import (
	"errors"

	"go.uber.org/zap"
)

type PokdengGame struct {
	players             []*PokdengPlayer
	firstTurnPlayer     *PokdengPlayer // Player that will take first turn.
	deck                Deck
	ActivePlayerIndex   int            // Index of active player in playerArray.
	ActivePlayer        *PokdengPlayer // Active player that can do an action (jua, geb, gerd, faak and ting) //.
	dealer              *PokdengPlayer
	dealerPosition      uint8
	botDealer           *PokdengPlayer
	DealerWinPlayer     []*PokdengPlayer
	DealerLosePlayer    []*PokdengPlayer
	DealerWinTotal      int64
	DealerLoseTotal     int64
	DealerChipGainTotal int64
	TurnQueue           []*PokdengPlayer
	waitTurnCh          chan bool
	isHavePlayerDealer  bool
	IsJub               bool
}

func newPokdengGame(room *PokdengRoom) *PokdengGame {
	return &PokdengGame{
		deck:                NewDeck(),
		players:             make([]*PokdengPlayer, 0),
		DealerWinPlayer:     make([]*PokdengPlayer, 0),
		DealerLosePlayer:    make([]*PokdengPlayer, 0),
		DealerWinTotal:      0,
		DealerLoseTotal:     0,
		DealerChipGainTotal: 0,
	}
}

func initializePokdengGame(game *PokdengGame) {
	game.deck = NewDeck()
	game.deck.shuffle()
	game.ActivePlayerIndex = 0
	game.ActivePlayer = &PokdengPlayer{}

	game.dealer = nil
	game.isHavePlayerDealer = false

	game.botDealer = NewPokdengPlayer()
	game.botDealer.UID = ""

	game.waitTurnCh = make(chan bool)
	// game.turnDirection = RIGHT
}

func (game *PokdengGame) setPlayerDealer(id string, player *PokdengPlayer) {
	game.dealer = player
	game.dealerPosition = player.SeatPosition
	game.isHavePlayerDealer = true
}

// func initializePokdengDealCard(room *PokdengRoom) {
// }

func initializePokdengPlayers(game *PokdengGame) {

	// for _, player := range game.players {
	// }
}

func (game *PokdengGame) dealCardToAllPlayer(room *PokdengRoom) {
	if game.dealer == nil {
		game.isHavePlayerDealer = false
	}
	numCardNeed := 2

	// Deal cards
	for indexCard := 0; indexCard < numCardNeed; indexCard++ {
		for _, player := range game.players {
			player.HandCardList.addCard(game.deck.deal())
			// zap.S().Debugf("Dummyroom < initializeDealCard: Player = %v Cards = %v", indexPlayer, game.players[indexPlayer].HandCardList)
		}

		if game.isHavePlayerDealer {
			game.dealer.HandCardList.addCard(game.deck.deal())
		} else {
			game.botDealer.HandCardList.addCard((game.deck.deal()))
		}
	}

}

func (game *PokdengGame) jua(player *PokdengPlayer) (Card, error) {
	zap.S().Debug("DummyGame < Jua")
	if game.ActivePlayer != nil && game.ActivePlayer.UID != player.UID {
		return Card{}, errors.New("the player is not active")
	}

	juaCard := game.deck.deal()

	if juaCard.id == 0 {
		return Card{}, errors.New("no more juaCard")
	}

	player.HandCardList.addCard(juaCard)
	return juaCard, nil
}

func (game *PokdengGame) botDealerJua() {
	juaCard := game.deck.deal()

	if juaCard.id == 0 {
		zap.S().Debug("DummyGame < No more juaCard")

		// return Card{}, errors.New("no more juaCard")
	}

	game.botDealer.HandCardList.addCard(juaCard)
}

func (game *PokdengGame) getDealer() *PokdengPlayer {
	if game.isHavePlayerDealer {
		return game.dealer
	}
	return game.botDealer
}

func (game *PokdengGame) getDealerSeatPosition() uint8 {
	if game.isHavePlayerDealer {
		return game.dealer.SeatPosition
	}
	return 0
}

func (game *PokdengGame) getPlayers() []*PokdengPlayer {
	return game.players
}

func (game *PokdengGame) isDealerBot() bool {
	if game.isHavePlayerDealer {
		return false
	}
	return true
}

func (game *PokdengGame) compareTwoPlayerWhoWin(p1 *PokdengPlayer, p2 *PokdengPlayer) (*PokdengPlayer, int, int) {
	p1CardResult, p1PointMultiply, p1Point := p1.GetCardResultAndDeng()
	p2CardResult, p2PointMultiply, p2Point := p2.GetCardResultAndDeng()

	if p1CardResult < p2CardResult {
		return p1, p1CardResult, p1PointMultiply
	} else if p2CardResult < p1CardResult {
		return p2, p2CardResult, p2PointMultiply
	} else {
		if p1CardResult == p2CardResult && (p1CardResult == POKNINE || p1CardResult == POKEIGHT) {
			return nil, 0, 1
		} else if p1CardResult == p2CardResult && (p1CardResult == TONG || p1CardResult == STRAIGHT) {
			p1Card := GetHighestCard(p1.HandCardList.cards)
			p2Card := GetHighestCard(p2.HandCardList.cards)

			if p1Card.number > p2Card.number {
				return p1, p1CardResult, p1PointMultiply
			} else if p1Card.number < p2Card.number {
				return p2, p2CardResult, p2PointMultiply
			} else {
				return nil, 0, 1
			}
		} else if p1CardResult == p2CardResult && p1CardResult == SIAN {
			return nil, SIAN, 1
		} else {
			if p1Point > p2Point {
				return p1, POINT, p1PointMultiply
			} else if p1Point < p2Point {
				return p2, POINT, p2PointMultiply
			} else {
				return nil, 0, 1
			}
		}
	}
}

// ProcessScoreEndGame Calculate chipGain
func (game *PokdengGame) ProcessScoreEndGame() {
	dealerPlayer := game.getDealer()
	game.DealerWinPlayer = make([]*PokdengPlayer, 0)
	game.DealerLosePlayer = make([]*PokdengPlayer, 0)

	for _, player := range game.players {
		player.playerState = PokdengPlayerState.EndGame

		//check point and add chip gain
		winPlayer, _, winMultiply := game.compareTwoPlayerWhoWin(dealerPlayer, player)

		if winPlayer == dealerPlayer {
			calculateChip := int64(player.Bet) * int64(winMultiply)
			dealerPlayer.ChipGain += calculateChip
			player.ChipGain -= calculateChip

			game.DealerWinPlayer = append(game.DealerWinPlayer, player)
			game.DealerWinTotal += calculateChip
		} else if winPlayer == player {
			calculateChip := int64(player.Bet) * int64(winMultiply)
			player.ChipGain += calculateChip
			dealerPlayer.ChipGain -= calculateChip

			game.DealerLosePlayer = append(game.DealerLosePlayer, player)
			game.DealerLoseTotal += calculateChip
		}

		game.DealerChipGainTotal = dealerPlayer.ChipGain
		// if !game.IsDealerBot() {
		// 	dealerPlayer.UpdateStatistic()
		// }
	}
}

func (game *PokdengGame) ProcessScoreWithJub(numJub int) {
	dealerPlayer := game.getDealer()
	game.DealerWinPlayer = make([]*PokdengPlayer, 0)
	game.DealerLosePlayer = make([]*PokdengPlayer, 0)
	for _, player := range game.players {
		if player.HandCardList.GetCardsCount() == numJub {
			player.IsJub = true
			//check point and add chip gain
			winPlayer, _, winMultiply := game.compareTwoPlayerWhoWin(dealerPlayer, player)
			if winPlayer == dealerPlayer {
				calculateChip := int64(player.Bet) * int64(winMultiply)
				dealerPlayer.ChipGain += calculateChip
				player.ChipGain -= calculateChip

				game.DealerWinPlayer = append(game.DealerWinPlayer, player)
				game.DealerWinTotal += calculateChip
			} else if winPlayer == player {
				calculateChip := int64(player.Bet) * int64(winMultiply)
				player.ChipGain += calculateChip
				dealerPlayer.ChipGain -= calculateChip
				game.DealerLosePlayer = append(game.DealerLosePlayer, player)
				game.DealerLoseTotal += calculateChip
			}

			game.DealerChipGainTotal = dealerPlayer.ChipGain
		}
	}
}

func (game *PokdengGame) ProcessScoreWithOther() {
	dealerPlayer := game.getDealer()
	for _, player := range game.players {
		if !player.IsJub {
			//check point and add chip gain
			winPlayer, _, winMultiply := game.compareTwoPlayerWhoWin(dealerPlayer, player)
			if winPlayer == dealerPlayer {
				calculateChip := int64(player.Bet) * int64(winMultiply)
				dealerPlayer.ChipGain += calculateChip
				player.ChipGain -= calculateChip

				game.DealerWinPlayer = append(game.DealerWinPlayer, player)
				game.DealerWinTotal += calculateChip
			} else if winPlayer == player {
				calculateChip := int64(player.Bet) * int64(winMultiply)
				player.ChipGain += calculateChip
				dealerPlayer.ChipGain -= calculateChip
				game.DealerLosePlayer = append(game.DealerLosePlayer, player)
				game.DealerLoseTotal += calculateChip
			}

			game.DealerChipGainTotal = dealerPlayer.ChipGain
		}
	}
}
