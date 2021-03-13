package game

import (
	"math/rand"
	"time"
)

// Deck is a deck that contain 52 cards
type Deck struct {
	cards []Card
}

func (d Deck) String() string {
	str := "Deck: [ "
	for _, card := range d.cards {
		str = str + card.getCardString() + " "
	}
	str = str + "]"

	return str
}

// NewDeck create new 52 card for using
func NewDeck() Deck {
	return Deck{
		cards: CreateCardsInDeck(),
	}
}

// @param bHasJokers (boolean) Create the deck with or with out jokers.
// @return A table contains Card objects.
func CreateCardsInDeck() []Card {
	var cards []Card

	for i := 0; i < 52; i++ {
		cardID := uint8(i + 1) // card id start with 1
		cards = append(cards, createCardByID(cardID))
	}

	return cards
}

// Shuffle the Deck.
func (d *Deck) shuffle() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(d.cards), func(i, j int) {
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	})
}

// Deal the card at top of the deck.
// @return A Card object.
func (d *Deck) deal() Card {
	var card Card
	if len(d.cards) > 0 {
		card, d.cards = d.cards[len(d.cards)-1], d.cards[:len(d.cards)-1]
		return card
	}
	return Card{}
}

// Number of cards left in the deck.
func (d *Deck) numCardLeft() int {
	return len(d.cards)
}
