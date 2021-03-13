package game

import (
	"errors"
	"math"
	"strings"

	"go.uber.org/zap"
)

// CardList wraps card slices for convenience use
type CardList struct {
	cards []Card
}

func (cardList CardList) String() string {
	cardListStr := "[ "

	for _, card := range cardList.cards {
		cardListStr = cardListStr + card.getCardString() + " "
	}

	cardListStr = cardListStr + "]"

	return cardListStr
}

// IsThreeOfAKind means same numbers for 3 cards
func (cardList CardList) IsThreeOfAKind() bool {
	if len(cardList.cards) != 3 {
		return false
	}

	return cardList.cards[0].number == cardList.cards[1].number &&
		cardList.cards[1].number == cardList.cards[2].number
}

// IsFourOfAKind means same numbers for 4 cards
func (cardList CardList) IsFourOfAKind() bool {
	if len(cardList.cards) != 4 {
		return false
	}

	return cardList.cards[0].number == cardList.cards[1].number &&
		cardList.cards[1].number == cardList.cards[2].number &&
		cardList.cards[2].number == cardList.cards[3].number
}

func (cardList CardList) isDuplicateCard() bool {
	length := len(cardList.cards)
	if length == 0 {
		return false
	}
	for index, card := range cardList.cards {
		for i := index + 1; i < len(cardList.cards); i++ {
			expectCard := cardList.cards[i]
			if card.id == expectCard.id {
				zap.L().Error("found duplicate card", zap.String("card", card.getCardString()))
				return true
			}
		}
	}
	return false
}

func convertNumberOneToFourteen(cards []Card) {
	for index := range cards {
		if cards[index].number == 1 {
			cards[index].number = 14
		}
	}
}

// IsSequence is check that sequence
func (cardList CardList) IsSequence() bool {
	if len(cardList.cards) < 3 {
		return false
	}

	var tempCards = make([]Card, len(cardList.cards))
	copy(tempCards, cardList.cards)

	convertNumberOneToFourteen(tempCards)

	OrderedBy(number).Sort(tempCards)

	for i := 0; i < len(tempCards); i++ {
		if i+1 == len(tempCards) {
			break
		}

		if !tempCards[i].isNextNum(tempCards[i+1]) {
			return false
		}
	}

	return true
}

// IsSameSuit is check the cardList that have one suit!
func (cardList CardList) IsSameSuit() (bool, error) {
	if len(cardList.cards) == 0 {
		return false, errors.New("must have at least 1 card")
	}

	var suit Suit
	var isFirstTime = true

	for _, card := range cardList.cards {
		if isFirstTime {
			suit = card.suit
			isFirstTime = false
			continue
		}

		if suit != card.suit {
			return false, nil
		}
	}

	return true, nil
}

// IsSameNumber is check the cardList that have one number!
func (cardList CardList) IsSameNumber() (bool, error) {
	if len(cardList.cards) == 0 {
		return false, errors.New("must have at least 1 card")
	}

	var number uint8
	var isFirstTime = true

	for _, card := range cardList.cards {
		if isFirstTime {
			number = card.number
			isFirstTime = false
			continue
		}

		if number != card.number {
			return false, nil
		}
	}

	return true, nil
}

// IsSequenceAndSameSuit as a name declared
func (cardList CardList) IsSequenceAndSameSuit() bool {
	isSameSuit, err := cardList.IsSameSuit()
	if err != nil {
		zap.L().Error("err in IsSequenceAndSameSuit", zap.Error(err))
	}

	if !isSameSuit {
		return false
	}

	return cardList.IsSequence()
}

// IsPok as a name declared
func (cardList CardList) IsPok() bool {
	if len(cardList.cards) > 2 {
		return false
	}

	point := cardList.GetPokdengCardPoint()
	if point == 8 || point == 9 {
		return true
	}
	return false
}

// IsSian as a name declared
func (cardList CardList) IsSian() bool {
	if len(cardList.cards) != 3 {
		return false
	}

	sian := true
	for _, v := range cardList.cards {
		if v.number != 11 && v.number != 12 && v.number != 13 {
			sian = false
		}
	}
	return sian
}

// GetPokdengCardPoint gets point from these cards
func (cardList CardList) GetPokdengCardPoint() int8 {
	point := 0
	for _, v := range cardList.cards {
		cardPoint := int(v.number)
		if cardPoint > 10 {
			cardPoint = 10
		}

		point += cardPoint
	}

	point = int(math.Mod(float64(point), float64(10)))
	return int8(point)
}

// GetHighestCard gets the most highest card from these cards
func GetHighestCard(cards []Card) Card {
	var highestCardID uint8
	var highestCard Card
	for i := len(cards) - 1; i >= 0; i-- {
		cardID := cards[i].id
		if cardID == 1 {
			cardID = 14
		}
		if highestCardID < cardID {
			highestCardID = cardID
			highestCard = cards[i]
		}
	}
	return highestCard
}

// Will return something like "3♥ 7♠ K♣".
func (cardList *CardList) getCardString() string {
	s := ""
	for _, v := range cardList.cards {
		s = s + " " + v.getCardString()
	}
	return strings.Trim(s, " ")
}

// Is there a given Card in this CardList.
// @param card A Card object.
// @return boolean.
func (cardList *CardList) hasCard(c Card) bool {
	for _, v := range cardList.cards {
		if c.id == v.id {
			return true
		}
	}
	return false
}

// Clone this card list to a new CardList object.
// @return A new CardList object that has same content.
func (cardList CardList) clone() CardList {
	cloneCardList := CardList{cards: make([]Card, len(cardList.cards))}
	copy(cloneCardList.cards, cardList.cards)
	return cloneCardList
}

// Add cards from a given cardList to this CardList. Will skip a duplicate card (not add).
// @param cardList A CardList object that it's content will be add to this CardList.
// The param cardList is not modified.
func (cardList *CardList) concat(cardList2 CardList) {
	for _, card := range cardList2.cards {
		if !cardList.hasCard(card) {
			cardList.addCard(card)
		}
	}
}

// Add a Card to this CardList.
// @param card A Card to add.
func (cardList *CardList) addCard(c Card) {
	cardList.cards = append(cardList.cards, c)
}

// Add several Cards to this CardList.
// @param ... A Cards to add in comma separate.
func (cardList *CardList) addCards(c []Card) {
	for _, card := range c {
		if !cardList.hasCard(card) {
			cardList.addCard(card)
		}
	}
}

// Add a Card by card ID.
// @param Card ID (number).
func (cardList *CardList) AddCardByID(i uint8) {
	cardList.cards = append(cardList.cards, createCardByID(i))
}

// Remove a Card from this CardList.
// @param card A Card to remove.
func (cardList *CardList) removeCard(c2 Card) {
	for i, c1 := range cardList.cards {
		if c1.id == c2.id {
			cardList.cards = append(cardList.cards[:i], cardList.cards[i+1:]...)
			return
		}
	}
}

// Remove all cards in the given cardList from this CardList.
// @param cardList A CardList that contains cards to remove.
func (cardList *CardList) removeCardByCardList(cardList2 CardList) {
	for _, c2 := range cardList2.cards {
		cardList.removeCard(c2)
	}
}

// Find index of a given card in this CardList.
// @param card A Card object that need to find index.
// @return Index of the card if found or 0 if not found.
func (cardList CardList) getIndexOfCard(c2 Card) (int, error) {
	for i, c1 := range cardList.cards {
		if c1 == c2 {
			return i, nil
		}
	}
	return 0, errors.New("index not found")
}

// Get the last card.
// @return A Card object.
func (cardList CardList) getLastCard() Card {
	return cardList.cards[len(cardList.cards)-1]
}

// GetCardsCount return length of cards in current cardlist.
func (cardList CardList) GetCardsCount() int {
	return len(cardList.cards)
}

// GetCardAt find card object at given index and then return.
func (cardList CardList) GetCardAt(index int) (Card, error) {
	if index < 0 || index >= len(cardList.cards) {
		return Card{}, errors.New("index out of range")
	}
	return cardList.cards[index], nil
}

// Get a CardList object that contains several Cards from a given index to index + (length-1) and also remove them from this CardList.
// Both index and index + (length-1) must be not out of range.
// @param index Start index to pop.
// @param length Number of Cards to pop. If not provide this method will pop all Cards start from the given index.
// @return A CardList object.
func (cardList *CardList) popCardFrom(index int, length int) (CardList, error) {
	if index >= len(cardList.cards) {
		return CardList{}, errors.New("index out of range")
	}

	if (index + length - 1) >= len(cardList.cards) {
		return CardList{}, errors.New("index + length - 1 is out of range")
	}

	var returnCardList CardList
	if length == 0 {
		returnCardList = CardList{cards: make([]Card, len(cardList.cards)-index)}
		copy(returnCardList.cards, cardList.cards[index:])
	} else {
		returnCardList = CardList{cards: make([]Card, length)}
		copy(returnCardList.cards, cardList.cards[index:length+index])
	}

	cardList.cards = cardList.cards[:index+copy(cardList.cards[index:], cardList.cards[length+index:])]
	return returnCardList, nil
}

// See that this CardList is a superset of the given cardList.
// @param cardList Another CardList object.
// @return boolean.
func (cardList *CardList) isSupersetOfCardList(cardList2 CardList) bool {
	for _, card := range cardList2.cards {
		if !cardList.hasCard(card) {
			return false
		}
	}
	return true
}

func (cardList CardList) GetCards() []Card {
	return cardList.cards
}
