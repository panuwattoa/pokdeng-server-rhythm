package game

import (
	"math"
	"sort"
	"strconv"
	"sync"
)

// Card is structure suit and number
type Card struct {
	id     uint8
	suit   Suit
	number uint8
}

func (c Card) String() string {
	// return fmt.Sprintf("{id: %d, suit: %d, number: %d}", c.id, uint8(c.suit), c.number)
	return c.getCardString()
}

type lessFunc func(p1, p2 *Card) bool

type multiSorter struct {
	cards []Card
	less  []lessFunc
}

var number = func(c, c2 *Card) bool {
	if c.number == 1 {
		return 14 < c2.number
	}

	if c2.number == 1 {
		return c.number < 14
	}
	return c.number < c2.number
}

var suit = func(c, c2 *Card) bool {
	return c.suit < c2.suit
}

// Suit of card
type Suit uint8

// Suit number
const (
	CLUBS    Suit = 1
	DIAMONDS Suit = 2
	HEARTS   Suit = 3
	SPADES   Suit = 4
)

var gCards = &sync.Map{}

func (ms *multiSorter) Sort(cards []Card) {
	ms.cards = cards
	sort.Sort(ms)
}

func OrderedBy(less ...lessFunc) *multiSorter {
	return &multiSorter{
		less: less,
	}
}

func (ms *multiSorter) Len() int {
	return len(ms.cards)
}

func (ms *multiSorter) Swap(i, j int) {
	ms.cards[i], ms.cards[j] = ms.cards[j], ms.cards[i]
}

func (ms *multiSorter) Less(i, j int) bool {
	p, q := &ms.cards[i], &ms.cards[j]
	var k int
	for k = 0; k < len(ms.less)-1; k++ {
		less := ms.less[k]
		switch {
		case less(p, q):
			return true
		case less(q, p):
			return false
		}
	}

	return ms.less[k](p, q)
}

func NewCard(_number uint8, suit Suit) Card {
	return Card{
		number: _number,
		suit:   suit,
		id:     getIDFromNumberAndSuit(_number, suit),
	}
}

func getIDFromNumberAndSuit(number uint8, suit Suit) uint8 {
	return ((number - 1) * 4) + uint8(suit)
}

func getNumFromID(id uint8) uint8 {
	return uint8(math.Ceil(float64(id) / 4.0))
}

func getSuitFromID(id uint8) Suit {
	return Suit(id - (getNumFromID(id)-1)*4)
}

// GetCardNumber return current card number.
func (c Card) GetCardNumber() uint8 {
	return c.number
}

// GetCardID return current card id.
func (c Card) GetCardID() uint8 {
	return c.id
}

// GetCardSuit return current card suit.
func (c Card) GetCardSuit() Suit {
	return c.suit
}

func (c Card) getCardString() string {
	cardString := ""
	switch c.number {
	case 1:
		cardString = "A"
	case 11:
		cardString = "J"
	case 12:
		cardString = "Q"
	case 13:
		cardString = "K"
	default:
		cardString = strconv.Itoa(int(c.number))
	}

	switch c.suit {
	case CLUBS:
		cardString = cardString + "♣"
	case DIAMONDS:
		cardString = cardString + "♦"
	case HEARTS:
		cardString = cardString + "♥"
	case SPADES:
		cardString = cardString + "♠"
	}

	if cardString == "0" {
		cardString = "None"
	}

	return cardString
}

// func (c Card) isEqual(c2 Card) bool {
// 	return reflect.DeepEqual(c, c2)
// }

func (c Card) isSameNum(c2 Card) bool {
	return c.number == c2.number
}

func (c Card) isSameSuit(c2 Card) bool {
	return c.suit == c2.suit
}

func (c Card) isNextNum(c2 Card) bool {
	if c.number == 14 {
		c.number = 1
	}
	if c2.number == 14 {
		c2.number = 1
	}

	nextNumber := (c.number + 1)
	if nextNumber == 14 {
		nextNumber = nextNumber - 13
	}

	return nextNumber == c2.number
}

func (c Card) isNextNumSameSuit(c2 Card) bool {
	if !c.isSameSuit(c2) {
		return false
	}

	return c.isNextNum(c2)
}

func (c Card) isPreviousNumSameSuit(c2 Card) bool {
	if !c.isSameSuit(c2) {
		return false
	}

	nextNumber := (c2.number + 1)
	if nextNumber == 14 {
		nextNumber = nextNumber - 13
	}
	return c.number == nextNumber
}

func (c Card) isNext2NumSameSuit(c2 Card) bool {
	if !c.isSameSuit(c2) {
		return false
	}

	var nextNum uint8 = c.number + 2
	if nextNum > 13 {
		nextNum = nextNum % 13
	}

	return nextNum == c2.number
}

func (c Card) isPrevious2NumSameSuit(c2 Card) bool {
	if !c.isSameSuit(c2) {
		return false
	}

	var nextNum uint8 = c2.number + 2
	if nextNum > 13 {
		nextNum = nextNum % 13
	}

	return nextNum == c.number
}

func createCardByID(id uint8) Card {
	card, ok := gCards.Load(id)

	if !ok {
		number := getNumFromID(id)
		suit := getSuitFromID(id)
		card := NewCard(number, suit)
		gCards.Store(id, card)
		return card
	}

	return card.(Card)
}

func createCardByNumAndSuit(number uint8, suit Suit) Card {
	id := getIDFromNumberAndSuit(number, suit)
	card, ok := gCards.Load(id)
	if !ok {
		card = NewCard(number, suit)
		gCards.Store(id, card)
	}
	return card.(Card)
}

func createNextNumSameSuitCard(c Card) Card {
	nextNum := c.number + 1
	if nextNum == 14 {
		nextNum = 1
	}
	return createCardByNumAndSuit(nextNum, c.suit)
}

func createNext2NumSameSuitCard(c Card) Card {
	nextNum := c.number + 2
	if nextNum > 13 {
		nextNum = nextNum - 13
	}
	return createCardByNumAndSuit(nextNum, c.suit)
}

func createPreviousNumSameSuitCard(c Card) Card {
	previousNum := c.number - 1
	if previousNum == 0 {
		previousNum = 13
	}
	return createCardByNumAndSuit(previousNum, c.suit)
}

func createPrevious2NumSameSuitCard(c Card) Card {
	previousNum := c.number - 2
	if previousNum < 1 {
		previousNum = previousNum + 13
	}
	return createCardByNumAndSuit(previousNum, c.suit)
}

// // CompareTwoCardNumber is c1 > c2 return 1 elseif c1 < c2 return 2 else return 0
// func CompareTwoCardNumber(c1 Card, c2 Card) uint8 {
// 	tempC1 := NewCard(c1.number, c1.suit)
// 	tempC2 := NewCard(c2.number, c2.suit)

// 	if tempC1.number == 1 {
// 		tempC1.number = 14
// 	}
// 	if tempC2.number == 1 {
// 		tempC2.number = 14
// 	}

// 	if tempC1.number > tempC2.number {
// 		return 1
// 	} else if tempC1.number < tempC2.number {
// 		return 2
// 	} else {
// 		return 0
// 	}
// }
