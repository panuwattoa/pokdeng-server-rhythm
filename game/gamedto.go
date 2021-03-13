package game

import "encoding/json"

type JoinDto struct {
	NumSitUser  uint8
	BetRate     uint64
	MaxBetRate  uint64
	UserSit     []UserSitDto
	RoomPlaying bool
}

type UserSitDto struct {
	UID        string
	Name       string
	Position   uint8
	ProfilePic string
}

type RoomDealer struct {
	Position        uint8
	IsDealerBot     bool
	DealerTurnCount uint8
	DealerName      string
}

type PlayerBetDto struct {
	Position uint8
	BetRate  uint64
}

type PlayerLeave struct {
	Position uint8
}

type ShowCardSto struct {
	Position      uint8
	CardResult    uint8
	PointMultiply uint8
	Point         uint8
	HandCard      []HandCard
	PlayerList    []string
}

type HandCard struct {
	CardID uint8
}

type ChangeTurnDto struct {
	UID          string
	Position     uint8
	TurnTime     uint16
	IsDealerTurn bool
}

type PlayerJuaDto struct {
	UID           string
	Position      uint8
	CardID        uint8
	CardResult    uint8
	PointMultiply uint8
	Point         uint8
}

type UserBet struct {
	Bet int `json:"bet"`
}
type PlayGameResult struct {
	Result      int
	Chip        int64
	CurrentChip int64
}

type Leave struct {
	UID string `json:"UID"`
}
type Jub struct {
	Jub int `json:"jub"`
}

type DelerResultDto struct {
	DealerWinName      []string
	DealerLostName     []string
	DealerWinTotal     int
	DealerLostTotal    int
	DealerChipGenTotal int
	CurrentChip        int
}

func CreateDealerResult(dto DelerResultDto) []byte {
	data, _ := json.Marshal(dto)
	return data
}

func CreateResult(result int, chip int64, Cchip int64) []byte {
	data, _ := json.Marshal(PlayGameResult{
		Result:      result,
		Chip:        chip,
		CurrentChip: Cchip,
	})
	return data
}
func CreatePlayerLeave(position uint8) []byte {
	data, _ := json.Marshal(PlayerLeave{
		Position: position,
	})
	return data
}
func (m *MatchState) CreateJoinDto() []byte {
	var userSit []UserSitDto
	for id, seat := range m.sitUser {
		userSit = append(userSit, UserSitDto{
			UID:      id,
			Name:     m.UserName[id],
			Position: seat.Position,
		})
	}
	joinData := JoinDto{
		NumSitUser:  uint8(len(m.sitUser)),
		BetRate:     m.BetRate,
		MaxBetRate:  m.MaxBetRate,
		UserSit:     userSit,
		RoomPlaying: m.roomPlaying,
	}

	data, _ := json.Marshal(joinData)
	return data
}

func CreateOtherJoin(id string, name string, position uint8) []byte {
	userSit := UserSitDto{
		UID:      id,
		Name:     name,
		Position: position,
	}
	data, _ := json.Marshal(userSit)
	return data
}

func CreateSetDealer(position uint8, isDealerBot bool, dealerTurnCOunt uint8, dealerName string) []byte {
	jsonData := RoomDealer{
		Position:        position,
		IsDealerBot:     isDealerBot,
		DealerTurnCount: dealerTurnCOunt,
		DealerName:      dealerName,
	}
	data, _ := json.Marshal(jsonData)
	return data
}

func CreatePlayerBet(position uint8, bet uint64) []byte {
	jsonData := PlayerBetDto{
		Position: position,
		BetRate:  bet,
	}
	data, _ := json.Marshal(jsonData)
	return data
}

func (player *PokdengPlayer) CreatePokCard() []byte {
	cardResult, pointMultiply, point := player.GetCardResultAndDeng()
	var handCard []HandCard

	for _, card := range player.HandCardList.GetCards() {
		handCard = append(handCard, HandCard{
			card.GetCardID(),
		})
	}

	jsonData := ShowCardSto{
		Position:      player.SeatPosition,
		CardResult:    uint8(cardResult),
		PointMultiply: uint8(pointMultiply),
		Point:         uint8(point),
		HandCard:      handCard,
	}
	data, _ := json.Marshal(jsonData)
	return data
}

func (player *PokdengPlayer) CreateDealCard(players []string) []byte {
	cardResult, pointMultiply, point := player.GetCardResultAndDeng()
	var handCard []HandCard

	for _, card := range player.HandCardList.GetCards() {
		handCard = append(handCard, HandCard{
			card.GetCardID(),
		})
	}

	jsonData := ShowCardSto{
		Position:      player.SeatPosition,
		CardResult:    uint8(cardResult),
		PointMultiply: uint8(pointMultiply),
		Point:         uint8(point),
		HandCard:      handCard,
		PlayerList:    players,
	}
	data, _ := json.Marshal(jsonData)
	return data
}
func CreateChangeTurn(id string, position uint8, time uint16, isDealerTurn bool) []byte {
	jsonData := ChangeTurnDto{
		UID:          id,
		Position:     position,
		TurnTime:     time,
		IsDealerTurn: isDealerTurn,
	}
	data, _ := json.Marshal(jsonData)
	return data
}

func (player *PokdengPlayer) CreateJua(isShow bool, cardID uint8) []byte {
	var cardResult, pointMultiply int
	var point int8
	if isShow {
		cardResult, pointMultiply, point = player.GetCardResultAndDeng()
	}
	jsonData := PlayerJuaDto{
		UID:           player.UID,
		Position:      player.SeatPosition,
		CardID:        cardID,
		CardResult:    uint8(cardResult),
		PointMultiply: uint8(pointMultiply),
		Point:         uint8(point),
	}
	data, _ := json.Marshal(jsonData)
	return data
}
