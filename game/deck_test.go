package game

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateCardsInDeck(t *testing.T) {
	t.Parallel()
	cards := CreateCardsInDeck()

	assert.Equal(t, len(cards), 52)
	assert.Equal(t, cards[8].number, uint8(3))
}

func TestShuffle(t *testing.T) {
	t.Parallel()
	sortDeck := NewDeck()
	deck := NewDeck()

	deck.shuffle()

	assert.False(t, reflect.DeepEqual(sortDeck, deck))
}

func TestShuffleMustBeNotTheSame(t *testing.T) {
	t.Parallel()
	deck1 := NewDeck()
	deck2 := NewDeck()

	deck1.shuffle()
	deck2.shuffle()

	assert.False(t, reflect.DeepEqual(deck1, deck2))
}

func TestDeal(t *testing.T) {
	t.Parallel()
	deck := NewDeck()

	assert.Equal(t, len(deck.cards), 52)

	deck.deal()

	assert.Equal(t, len(deck.cards), 51)

}
