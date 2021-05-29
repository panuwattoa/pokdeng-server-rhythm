package game

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"
	"math/rand"
	"pokdeng-server/handler"
	"strconv"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
)

const (
	TickRate         = 1
	InitialJoinTicks = 60
)

type MatchState struct {
	debug               bool
	presences           map[string]runtime.Presence
	seatsCh             chan *Seat
	sitUser             map[string]*Seat
	BetRate             uint64
	MaxBetRate          uint64
	roomPlaying         bool
	betCount            int
	roomGameOpCode      int
	UserName            map[string]string
	RequestLeave        map[string]string
	RequestCancelDealer map[string]string
	timerRoom           *time.Timer
}

type Seat struct {
	// Position is a seat position that have 1, 2, 3, 4 possible value
	Position uint8
}

type PokdengRoom struct {
	ArrayRandomDealer []string
	DealerTurnCount   int8
	DealerID          *string
	PokdengPlayer     map[string]*PokdengPlayer
	game              *PokdengGame
	turnTime          time.Time
	tax               float64
}

// MatchInit init room
func (room *PokdengRoom) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	var debug bool
	if d, ok := params["debug"]; ok {
		debug, _ = d.(bool)
	}
	label := params["gold"].(string)
	i64, _ := strconv.ParseInt(label, 10, 32)

	state := &MatchState{
		debug:               debug,
		presences:           make(map[string]runtime.Presence),
		seatsCh:             make(chan *Seat, 7),
		sitUser:             make(map[string]*Seat),
		BetRate:             uint64(i64),
		MaxBetRate:          uint64(i64) * 10,
		roomPlaying:         false,
		UserName:            make(map[string]string),
		RequestLeave:        make(map[string]string),
		RequestCancelDealer: make(map[string]string),
	}

	for i := 1; i <= 7; i++ {
		state.seatsCh <- &Seat{
			Position: uint8(i),
		}
	}
	room.tax = 1.5
	tickRate := 1
	room.PokdengPlayer = make(map[string]*PokdengPlayer)
	return state, tickRate, label
}

// MatchJoinAttempt attemp join
func (room *PokdengRoom) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	ableToJoin := true
	mState, _ := state.(*MatchState)
	reason := ""
	if len(mState.presences) >= 7 {
		ableToJoin = false
		reason = "ห้องเต็ม กรุณาเข้าใหม่อีกครั้ง"
	}
	if _, ok := mState.presences[presence.GetUserId()]; ok {
		ableToJoin = false
		reason = "นายท่านกำลังเล่นในห้องนี้กรุณนารอสักครู่"
	}
	logger.Debug("MaxBetRate ", mState.MaxBetRate)

	logger.Debug("len(mState.seatsCh) ", len(mState.seatsCh))

	if len(mState.seatsCh) == 0 {
		ableToJoin = false
		reason = "ห้องเต็ม กรุณาเข้าใหม่อีกครั้ง"
	}

	var wallet = room.GetAccountWallet(ctx, logger, nk, presence.GetUserId())
	if wallet < float64(mState.BetRate*5) {
		ableToJoin = false
		reason = "เงินไม่เพียงพอ"
	}
	return state, ableToJoin, reason
}

// MatchJoin join what happend when user  join room
func (room *PokdengRoom) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	mState, _ := state.(*MatchState)
	var users []string
	for _, p := range presences {
		if _, ok := mState.presences[p.GetUserId()]; ok {
			return mState
		}
		if len(mState.seatsCh) == 0 {
			return mState
		}
		seat := func() *Seat {
			for seat := range mState.seatsCh {
				return seat
			}
			return nil
		}
		users = append(users, p.GetUserId())
		mState.presences[p.GetUserId()] = p
		if _, ok := mState.sitUser[p.GetUserId()]; !ok {
			mState.sitUser[p.GetUserId()] = seat()
			logger.Debug("Got seat !!!! ")
		} else {
			logger.Debug("Haveeeeeee.")
		}

		delete(mState.RequestLeave, p.GetUserId())

		if len(mState.presences) == 1 {
			mState.roomGameOpCode = GameOpCodeWaitOpenTable
		}

	}
	if users, err := nk.UsersGetId(ctx, users, nil); err != nil {
		// Handle error.
	} else {

		for _, u := range users {
			mState.UserName[u.Id] = u.DisplayName
		}
	}
	return mState
}

// MatchLeave leave
func (room *PokdengRoom) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	mState, _ := state.(*MatchState)
	for _, p := range presences {
		mState.RequestLeave[p.GetUserId()] = p.GetUserId()
	}
	if len(mState.presences) == 0 {
		return nil
	}
	return mState
}

// MatchLoop game loop
func (room *PokdengRoom) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	mState, _ := state.(*MatchState)

	// =========================================================================   Match Loop =============================================================
	for _, message := range messages {
		var sendMessage []byte
		var presence []runtime.Presence
		var err error

		switch message.GetOpCode() {
		case OpCodeJoinRequest:
			for id, presend := range mState.presences {
				if id != message.GetUserId() {
					joiner, ok := mState.sitUser[message.GetUserId()]
					if ok {
						sendto := []runtime.Presence{}
						sendto = append(sendto, presend)
						sendMessage = CreateOtherJoin(message.GetUserId(), mState.UserName[message.GetUserId()], joiner.Position)
						err = dispatcher.BroadcastMessage(OpCodeOtherJoin, sendMessage, sendto, nil, true)
					} else {
						logger.Debug("=== == = = = Not found  !!!! == = = = = = = ")
					}
				}
			}
			sendMessage = mState.CreateJoinDto()
			presence = append(presence, mState.presences[message.GetUserId()])
			err = dispatcher.BroadcastMessage(OpCodeJoinRequest, sendMessage, presence, message, true)
		case OpCodeUserBet:
			if room.game == nil {
				break
			}
			mState.betCount++
			var bet uint64
			var userBet UserBet

			err := json.Unmarshal(message.GetData(), &userBet)
			if err != nil {
				bet = mState.BetRate
			} else {
				bet = uint64(userBet.Bet)
			}
			var wallet = room.GetAccountWallet(ctx, logger, nk, message.GetUserId())
			if float64(bet) > wallet {
				bet = uint64(wallet)
			}
			if room == nil {
				break
			}
			if player, ok := room.PokdengPlayer[message.GetUserId()]; ok {
				player.Bet = bet
			} else {
				break
			}
			sendMessage = CreatePlayerBet(mState.sitUser[message.GetUserId()].Position, bet)
			err = dispatcher.BroadcastMessage(message.GetOpCode(), sendMessage, nil, message, true)
		case OpCodeJua:
			if room.game == nil {
				break
			}
			if room.DealerID != nil && message.GetUserId() == *room.DealerID {
				room.Jua(logger, room.game.getDealer(), dispatcher, mState)
				room.game.ActivePlayer = nil
				time.AfterFunc(3*time.Second, func() {
					mState.roomGameOpCode = GameProcessScore
				})
			} else {
				room.Jua(logger, room.PokdengPlayer[message.GetUserId()], dispatcher, mState)
				mState.roomGameOpCode = GameChangeTurn
			}
			break
		case OpCodeNotJua:
			if room.game == nil {
				break
			}
			if room.DealerID != nil && message.GetUserId() == *room.DealerID {
				room.game.ActivePlayer = nil
				mState.roomGameOpCode = GameProcessScore
			} else {
				mState.roomGameOpCode = GameChangeTurn
			}
			break
		case OpCodeLeave:
			// var leaveID Leave
			// err := json.Unmarshal(message.GetData(), &leaveID)
			// if err == nil {
			if mState.roomPlaying {
				mState.RequestLeave[message.GetUserId()] = message.GetUserId()
			} else {
				sitUser, ok := mState.sitUser[message.GetUserId()]
				if ok {
					mState.seatsCh <- sitUser
					sendMessage := CreatePlayerLeave(sitUser.Position)
					err := dispatcher.BroadcastMessage(OpCodeLeave, sendMessage, nil, nil, true)
					if err != nil {
						logger.Info("cause an error when broadcasting", err)
					}

				} else {
					sendMessage := CreatePlayerLeave(0)
					if present, ok := mState.presences[message.GetUserId()]; ok {
						sendto := []runtime.Presence{}
						sendto = append(sendto, present)
						err := dispatcher.BroadcastMessage(OpCodeSelfLeave, sendMessage, sendto, nil, true)
						if err != nil {
							logger.Info("cause an error when broadcasting", err)
						}
					}
				}
				for index, requestDealer := range room.ArrayRandomDealer {
					if requestDealer == message.GetUserId() {
						room.ArrayRandomDealer = append(room.ArrayRandomDealer[:index], room.ArrayRandomDealer[index+1:]...)
						break
					}
				}
				delete(mState.sitUser, message.GetUserId())
				delete(mState.presences, message.GetUserId())
				delete(mState.UserName, message.GetUserId())
			}
			// }
		case OpCodeRequestDealer:
			room.ArrayRandomDealer = append(room.ArrayRandomDealer, message.GetUserId())
		case OpCodeCancelDealer:
			mState.RequestCancelDealer[message.GetUserId()] = message.GetUserId()
			for index, requestDealer := range room.ArrayRandomDealer {
				if requestDealer == message.GetUserId() {
					room.ArrayRandomDealer = append(room.ArrayRandomDealer[:index], room.ArrayRandomDealer[index+1:]...)
					break
				}
			}
		default:
			break
		}

		if err != nil {
			logger.Info("cause an error when broadcasting", err)
		}
	}

	// if !mState.roomPlaying && mState.roomGameOpCode != GameOpCodeWaitOpenTable {
	// 	mState.roomGameOpCode = GameOpCodeWaitOpenTable
	// }
	//============================================================== Game State ====================================================================================
	switch mState.roomGameOpCode {
	case GameOpCodeWaitOpenTable:
		mState.roomGameOpCode = GameOpCodeNone
		err := dispatcher.BroadcastMessage(OpCodeOpeningTable, nil, nil, nil, true)
		if err != nil {
			logger.Info("cause an error when broadcasting", err)
		}
		mState.timerRoom = time.AfterFunc(3*time.Second, func() {
			for id, sit := range mState.sitUser {
				var wallet = room.GetAccountWallet(ctx, logger, nk, id)
				if wallet < float64(mState.BetRate*5) {
					if room.DealerID != nil {
						if *room.DealerID == id {
							room.DealerID = nil
							room.DealerTurnCount = 0
						}
					}
					sendMessage := CreatePlayerLeave(sit.Position)
					err := dispatcher.BroadcastMessage(OpCodeLeave, sendMessage, nil, nil, true)
					if err != nil {
						logger.Info("cause an error when broadcasting", err)
					}
					mState.seatsCh <- sit
					delete(mState.sitUser, id)
					delete(mState.presences, id)
					delete(mState.UserName, id)
					delete(mState.RequestLeave, id)
				}
			}
			// check dealer
			needBoardcastDealer := true
			if room.DealerID != nil {
				room.DealerTurnCount = room.DealerTurnCount - 1
				if room.DealerTurnCount > 0 && room.verifyDealerChip(ctx, logger, nk, *room.DealerID, mState.BetRate) {
					needBoardcastDealer = false
				}
			} else if len(mState.sitUser) < 2 {
				needBoardcastDealer = false
				room.DealerID = nil
			}
			for id := range mState.sitUser {
				sendto := []runtime.Presence{}
				present, ok := mState.presences[id]
				if ok {
					sendto = append(sendto, present)
					if needBoardcastDealer && room.verifyDealerChip(ctx, logger, nk, id, mState.BetRate) {
						err := dispatcher.BroadcastMessage(OpCodeOpenRequestDealer, nil, sendto, nil, true)
						if err != nil {
							logger.Info("cause an error when broadcasting", err)
						}
					}
				}
			}
			mState.roomGameOpCode = GamePlayerReadyPlay

		})
	case GamePlayerReadyPlay:
		// check user chip
		mState.roomGameOpCode = GameOpCodeNone
		mState.timerRoom = time.AfterFunc(2*time.Second, func() {
			mState.roomGameOpCode = GameInitGamePlay
		})
	case GameInitGamePlay:
		mState.roomPlaying = true
		dealerPosition := uint8(0)
		mState.betCount = 0
		if len(mState.sitUser) == 0 {
			mState.roomGameOpCode = GameOpCodeNone
			mState.roomPlaying = false
			return mState
		}
		room.game = newPokdengGame(room)
		initializePokdengGame(room.game)
		// find dealer
		if len(mState.sitUser) > 1 && len(room.ArrayRandomDealer) > 0 && room.DealerID == nil {
			randomIndex := rand.Intn(len(room.ArrayRandomDealer))
			userIDIsDealer := room.ArrayRandomDealer[randomIndex]
			room.DealerID = &userIDIsDealer
			room.DealerTurnCount = 3
		}

		if room.DealerID != nil {
			if seat, ok := mState.sitUser[*room.DealerID]; ok {
				dealerPosition = mState.sitUser[*room.DealerID].Position
				player := NewPokdengPlayer()
				player.UID = *room.DealerID
				player.SeatPosition = seat.Position
				room.game.setPlayerDealer(*room.DealerID, player)
			}
		}
		var dealerName string
		if !room.game.isDealerBot() {
			if room.DealerID != nil {
				dealerName = mState.UserName[*room.DealerID]
			}
		}
		err := dispatcher.BroadcastMessage(OpCodeSetDealer, CreateSetDealer(dealerPosition, room.game.isDealerBot(), uint8(room.DealerTurnCount), dealerName), nil, nil, true)
		if err != nil {
			logger.Info("cause an error when broadcasting", err)
		}

		for id, seat := range mState.sitUser {
			u, ok := mState.presences[id]
			if ok {
				var dealerID string
				if room.DealerID != nil {
					dealerID = *room.DealerID
				}
				if dealerID != u.GetUserId() {
					player := NewPokdengPlayer()
					player.UID = u.GetUserId()
					player.SeatPosition = seat.Position
					room.PokdengPlayer[u.GetUserId()] = player
					room.game.players = append(room.game.players, player)
				}
			}
		}
		room.ArrayRandomDealer = make([]string, 0)
		mState.roomGameOpCode = GameStartBet
	case GameStartBet:
		for _, player := range room.game.players {
			sendto := []runtime.Presence{}
			present, ok := mState.presences[player.UID]
			if ok {
				sendto = append(sendto, present)
				err := dispatcher.BroadcastMessage(OpCodeGameStartBet, nil, sendto, nil, true)
				if err != nil {
					logger.Info("cause an error when broadcasting", err)
				}
			}
		}

		mState.roomGameOpCode = GameOpCodeWaitingBet
	case GameOpCodeWaitingBet:
		needWaitBet := mState.betCount < len(room.game.players)
		if needWaitBet {
			// mState.roomGameOpCode = GameOpCodeNone
			// if mState.timerRoom == nil {
			mState.timerRoom = time.AfterFunc(10*time.Second, func() {
				// logger.Debug("time out bet ..")
				for _, player := range room.game.players {
					if player.Bet == 0 {
						player.Bet = mState.BetRate
						mState.betCount++
						sendMessage := CreatePlayerBet(player.SeatPosition, player.Bet)
						sendto := []runtime.Presence{}
						present, ok := mState.presences[player.UID]
						if ok {
							sendto = append(sendto, present)
							err := dispatcher.BroadcastMessage(OpCodeUserBet, sendMessage, sendto, nil, true)
							if err != nil {
								logger.Info("cause an error when broadcasting", err)
							}
						}
					}
				}
				//	mState.roomGameOpCode = GameOpCodeDealCard
			})
			// }
		} else {
			if mState.timerRoom != nil {
				mState.timerRoom.Stop()
			}
			mState.roomGameOpCode = GameOpCodeDealCard
		}
	case GameOpCodeDealCard:
		room.game.dealCardToAllPlayer(room)
		var payerList []string
		for _, p := range room.game.players {
			payerList = append(payerList, p.UID)
		}

		if room.DealerID != nil {
			payerList = append(payerList, *room.DealerID)
		}
		for id := range mState.sitUser {
			sendto := []runtime.Presence{}
			present, ok := mState.presences[id]
			if ok {
				sendto = append(sendto, present)
				var player *PokdengPlayer
				if room.DealerID != nil && *room.DealerID == id {
					player = room.game.dealer
				} else {
					player = room.PokdengPlayer[id]
				}
				if player != nil {
					sendMessage := player.CreateDealCard(payerList)
					err := dispatcher.BroadcastMessage(OpCodeDealCard, sendMessage, sendto, nil, true)
					if err != nil {
						logger.Info("cause an error when broadcasting", err)
					}
				}

			}
		}
		mState.roomGameOpCode = GameOpCodeNone

		time.AfterFunc(8*time.Second, func() {
			if !room.game.getDealer().IsPok() {
				for _, player := range room.game.players {
					if player.IsPok() {
						logger.Debug("player Pok !! ..")
						//show card
						//broadcast show card
						sendMessage := player.CreatePokCard()
						err := dispatcher.BroadcastMessage(OpCodePok, sendMessage, nil, nil, true)
						if err != nil {
							logger.Info("cause an error when broadcasting", err)
						}
					} else {
						room.game.TurnQueue = append(room.game.TurnQueue, player)
					}
				}
				mState.roomGameOpCode = GameChangeTurn
			} else {
				logger.Debug("dealer Pok !! ..")
				mState.roomGameOpCode = GameProcessScore
			}
		})
	case GameChangeTurn:
		room.game.ActivePlayer = nil
		if len(room.game.TurnQueue) > 0 {
			mState.roomGameOpCode = GameOpCodeNone
			player := room.game.TurnQueue[0]
			room.game.TurnQueue = room.game.TurnQueue[1:]
			room.game.ActivePlayer = player
			room.turnTime = time.Now()
			cardPoint := player.HandCardList.GetPokdengCardPoint()
			if cardPoint < 4 {
				room.Jua(logger, player, dispatcher, mState)
				mState.roomGameOpCode = GameChangeTurn
				room.game.ActivePlayer = nil
				break
			}

			time.AfterFunc(10*time.Second, func() {
				if room.game.ActivePlayer == nil {
					return
				}
				if player.UID != room.game.ActivePlayer.UID {
					return
				}
				mState.roomGameOpCode = GameChangeTurn
				// room.Broadcast(packetMakerClient.MakePokdengEndTurn())
				// player.User.Remote.SendPacketWithWriter(packetMakerClient.MakePokdengAIPlay())

				//check card point below 4 juacard if not end turn
				cardPoint := player.HandCardList.GetPokdengCardPoint()
				if cardPoint < 4 {
					room.Jua(logger, player, dispatcher, mState)
				}
				room.game.ActivePlayer = nil
			})
			// //broadcast change turn
			err := dispatcher.BroadcastMessage(OpCodeChnageTurn, CreateChangeTurn(player.UID, player.SeatPosition, uint16(10-time.Since(room.turnTime).Seconds()), false), nil, nil, true)
			if err != nil {
				logger.Info("cause an error when broadcasting", err)
			}
		} else {
			mState.roomGameOpCode = GameOpCodeDealerPlay
		}
	case GameOpCodeDealerPlay:
		logger.Debug("GameOpCodeDealerPlay")
		if room.game.isDealerBot() {
			logger.Debug("dealer bot play..")
			cardPoint := room.game.botDealer.HandCardList.GetPokdengCardPoint()
			if cardPoint < 4 {
				room.game.botDealerJua()
				//broadcast jua card
				err := dispatcher.BroadcastMessage(OpCodeBotJua, []byte{}, nil, nil, true)
				if err != nil {
					logger.Info("cause an error when broadcasting", err)
				}
			}
			mState.roomGameOpCode = GameProcessScore
		} else {
			mState.roomGameOpCode = GameOpCodeNone
			// room.game.TurnQueue = room.game.TurnQueue[1:]
			room.game.ActivePlayer = room.game.getDealer()
			room.turnTime = time.Now()
			// //broadcast change turn
			err := dispatcher.BroadcastMessage(OpCodeChnageTurn, CreateChangeTurn(room.game.getDealer().UID, room.game.getDealer().SeatPosition, uint16(10-time.Since(room.turnTime).Seconds()), true), nil, nil, true)
			if err != nil {
				logger.Info("cause an error when broadcasting", err)
			}
			time.AfterFunc(10*time.Second, func() {
				if room.game.getDealer().IsJub {
					return
				}
				if room.game.ActivePlayer == nil {
					return
				}
				if room.game.getDealer().UID != room.game.ActivePlayer.UID {
					return
				}
				//check card point below 4 juacard if not end turn
				cardPoint := room.game.getDealer().HandCardList.GetPokdengCardPoint()
				if cardPoint < 4 {
					room.Jua(logger, room.game.getDealer(), dispatcher, mState)
				}
				room.game.ActivePlayer = nil
				mState.roomGameOpCode = GameProcessScore
			})
		}
	case GameProcessScore:
		mState.roomGameOpCode = GameOpCodeNone
		for _, player := range room.game.players {
			sendMessage := player.CreatePokCard()
			err := dispatcher.BroadcastMessage(OpCodePok, sendMessage, nil, nil, true)
			if err != nil {
				logger.Info("cause an error when broadcasting", err)
			}
		}
		room.game.ProcessScoreEndGame()
		room.saveResult(ctx, logger, nk)
		for _, player := range room.game.players {
			go func(player *PokdengPlayer) {
				if room != nil && room.game != nil {
					handler.SaveLogPlay(handler.LogPlay{
						UID:        player.UID,
						Name:       mState.UserName[player.UID],
						Bet:        int(mState.BetRate),
						MyCard:     player.HandCardList.getCardString(),
						DealerCard: room.game.getDealer().HandCardList.getCardString(),
						IsDealer:   room.game.isHavePlayerDealer,
						ChipGain:   int(player.ChipGain),
						DealerUID:  room.game.getDealer().UID,
					})
				}
			}(player)
		}

		go func() {
			if room.game.isHavePlayerDealer {
				if room.game.getDealer() != nil {
					handler.SaveLogPlay(handler.LogPlay{
						UID:        room.game.getDealer().UID,
						Name:       mState.UserName[room.game.getDealer().UID],
						Bet:        int(mState.BetRate),
						MyCard:     room.game.getDealer().HandCardList.getCardString(),
						DealerCard: room.game.getDealer().HandCardList.getCardString(),
						IsDealer:   room.game.isHavePlayerDealer,
						ChipGain:   int(room.game.getDealer().ChipGain),
						DealerUID:  room.game.getDealer().UID,
					})
				}
			}

		}()
		for _, player := range room.game.players {
			if player.ChipGain > 0 {
				sendto := []runtime.Presence{}
				present, ok := mState.presences[player.UID]
				if ok {
					sendto = append(sendto, present)
					gold := room.GetAccountWallet(ctx, logger, nk, player.UID)
					err := dispatcher.BroadcastMessage(OpCodeGameResult, CreateResult(1, player.ChipGain, int64(gold)), sendto, nil, true)
					if err != nil {
						logger.Info("cause an error when broadcasting", err)
					}
				}

			} else if player.ChipGain < 0 {
				sendto := []runtime.Presence{}
				present, ok := mState.presences[player.UID]
				if ok {
					sendto = append(sendto, present)
					gold := room.GetAccountWallet(ctx, logger, nk, player.UID)
					err := dispatcher.BroadcastMessage(OpCodeGameResult, CreateResult(2, player.ChipGain, int64(gold)), sendto, nil, true)
					if err != nil {
						logger.Info("cause an error when broadcasting", err)
					}
				}
			} else {
				sendto := []runtime.Presence{}
				present, ok := mState.presences[player.UID]
				if ok {
					sendto = append(sendto, present)
					gold := room.GetAccountWallet(ctx, logger, nk, player.UID)
					err := dispatcher.BroadcastMessage(OpCodeGameResult, CreateResult(3, player.ChipGain, int64(gold)), sendto, nil, true)
					if err != nil {
						logger.Info("cause an error when broadcasting", err)
					}
				}
			}
		}
		if room.game.isDealerBot() {
			sendMessage := room.game.botDealer.CreatePokCard()
			go handler.SaveLogDealer(handler.LogDealerPlay{
				BetRate: int(mState.BetRate),
				Value:   int(room.game.botDealer.ChipGain),
			})
			err := dispatcher.BroadcastMessage(OpBotDealerCard, sendMessage, nil, nil, true)
			if err != nil {
				logger.Info("cause an error when broadcasting", err)
			}
		} else {
			// todo:  send result to dealer
			sendMessage := room.game.dealer.CreatePokCard()
			err := dispatcher.BroadcastMessage(OpCodePok, sendMessage, nil, nil, true)
			if err != nil {
				logger.Info("cause an error when broadcasting", err)
			}
			sendto := []runtime.Presence{}
			present, ok := mState.presences[*room.DealerID]
			if ok {
				var winName []string
				var lostName []string
				for _, player := range room.game.DealerWinPlayer {
					if user, ok := mState.UserName[player.UID]; ok {
						winName = append(winName, user)
					}
				}
				for _, player := range room.game.DealerLosePlayer {
					if user, ok := mState.UserName[player.UID]; ok {
						lostName = append(lostName, user)
					}
				}
				sendto = append(sendto, present)
				gold := room.GetAccountWallet(ctx, logger, nk, *room.DealerID)
				err := dispatcher.BroadcastMessage(OpCodeDealerResult, CreateDealerResult(DelerResultDto{
					DealerWinTotal:     int(room.game.DealerWinTotal),
					DealerLostTotal:    int(room.game.DealerLoseTotal),
					DealerChipGenTotal: int(room.game.dealer.ChipGain),
					DealerWinName:      winName,
					DealerLostName:     lostName,
					CurrentChip:        int(gold),
				}), sendto, nil, true)
				if err != nil {
					logger.Info("cause an error when broadcasting", err)
				}
			}

		}

		time.AfterFunc(10*time.Second, func() {
			mState.roomGameOpCode = GameClearTable
		})
	case GameProcessScoreWithJub:
		mState.roomGameOpCode = GameOpCodeNone
		room.saveResultJub(ctx, logger, nk)
		for _, player := range room.game.players {
			go func(player *PokdengPlayer) {
				if room != nil && room.game != nil {
					handler.SaveLogPlay(handler.LogPlay{
						UID:        player.UID,
						Name:       mState.UserName[player.UID],
						Bet:        int(mState.BetRate),
						MyCard:     player.HandCardList.getCardString(),
						DealerCard: room.game.getDealer().HandCardList.getCardString(),
						IsDealer:   room.game.isHavePlayerDealer,
						ChipGain:   int(player.ChipGain),
						DealerUID:  room.game.getDealer().UID,
					})
				}
			}(player)
		}

		go func() {
			if room.game.isHavePlayerDealer {
				if room.game.getDealer() != nil {
					handler.SaveLogPlay(handler.LogPlay{
						UID:        room.game.getDealer().UID,
						Name:       mState.UserName[room.game.getDealer().UID],
						Bet:        int(mState.BetRate),
						MyCard:     room.game.getDealer().HandCardList.getCardString(),
						DealerCard: room.game.getDealer().HandCardList.getCardString(),
						IsDealer:   room.game.isHavePlayerDealer,
						ChipGain:   int(room.game.getDealer().ChipGain),
						DealerUID:  room.game.getDealer().UID,
					})
				}
			}
		}()
		// deler
		// todo:  send result to dealer
		sendMessage := room.game.dealer.CreatePokCard()
		err := dispatcher.BroadcastMessage(OpCodePok, sendMessage, nil, nil, true)
		if err != nil {
			logger.Info("cause an error when broadcasting", err)
		}

		sendto := []runtime.Presence{}
		present, ok := mState.presences[*room.DealerID]
		if ok {
			var winName []string
			var lostName []string
			for _, player := range room.game.DealerWinPlayer {
				if user, ok := mState.UserName[player.UID]; ok {
					winName = append(winName, user)
				}
			}
			for _, player := range room.game.DealerLosePlayer {
				if user, ok := mState.UserName[player.UID]; ok {
					lostName = append(lostName, user)
				}
			}
			sendto = append(sendto, present)
			gold := room.GetAccountWallet(ctx, logger, nk, *room.DealerID)
			err := dispatcher.BroadcastMessage(OpCodeDealerResult, CreateDealerResult(DelerResultDto{
				DealerWinTotal:     int(room.game.DealerWinTotal),
				DealerLostTotal:    int(room.game.DealerLoseTotal),
				DealerChipGenTotal: int(room.game.dealer.ChipGain),
				DealerWinName:      winName,
				DealerLostName:     lostName,
				CurrentChip:        int(gold),
			}), sendto, nil, true)
			if err != nil {
				logger.Info("cause an error when broadcasting", err)
			}
		}

		time.AfterFunc(10*time.Second, func() {
			mState.roomGameOpCode = GameClearTable
		})
	case GameClearTable:
		mState.roomGameOpCode = GameOpCodeNone
		logger.Debug("Clear Table..")
		err := dispatcher.BroadcastMessage(OpCodeClearTable, []byte{}, nil, nil, true)
		if err != nil {
			logger.Info("cause an error when broadcasting", err)
		}
		room.game = nil
		mState.roomPlaying = false
		for id := range mState.sitUser {
			var gold map[string]interface{}
			account, err := nk.AccountGetId(ctx, id)
			if err != nil {
				// Handle error.
			} else {
				err := json.Unmarshal([]byte(account.Wallet), &gold)
				if err == nil {
					if gold["gold"].(float64) < float64(mState.BetRate*5) {
						logger.Info("Request leave : %v  dcsdcs %v ", gold["gold"].(float64), float64(mState.BetRate*5))
						mState.RequestLeave[id] = id
					}
				}
			}

		}
		//leave
		for id := range mState.RequestLeave {
			sitUser, ok := mState.sitUser[id]
			if ok {
				sendMessage := CreatePlayerLeave(sitUser.Position)
				err := dispatcher.BroadcastMessage(OpCodeLeave, sendMessage, nil, nil, true)
				if err != nil {
					logger.Info("cause an error when broadcasting", err)
				}
				mState.seatsCh <- sitUser
			}
			for index, requestDealer := range room.ArrayRandomDealer {
				if requestDealer == id {
					delete(mState.RequestCancelDealer, *room.DealerID)
					room.DealerID = nil
					room.DealerTurnCount = 0
					room.ArrayRandomDealer = append(room.ArrayRandomDealer[:index], room.ArrayRandomDealer[index+1:]...)
					break
				}
			}
			delete(mState.sitUser, id)
			delete(mState.presences, id)
			delete(mState.UserName, id)
			delete(mState.RequestLeave, id)
		}
		if room.DealerID != nil {
			if room.DealerTurnCount <= 1 {
				room.DealerID = nil
			}
			if _, ok := mState.RequestCancelDealer[*room.DealerID]; ok {
				delete(mState.RequestCancelDealer, *room.DealerID)
				room.DealerID = nil
			}
		}
		mState.roomGameOpCode = GameOpCodeWaitOpenTable
		if len(mState.presences) == 0 {
			mState.roomGameOpCode = GameOpCodeNone
		}
		mState.betCount = 0
	}

	return state
}

// MatchTerminate when room end
func (room *PokdengRoom) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	return state
}

func (room *PokdengRoom) Jua(logger runtime.Logger, player *PokdengPlayer, dispatcher runtime.MatchDispatcher, mState *MatchState) {
	//jua
	logger.Debug("Jua..")

	juaCard, err := room.game.jua(player)
	if err != nil {
		return
	}

	for id, presence := range mState.presences {
		sendto := []runtime.Presence{}
		if id == player.UID {
			sendto = append(sendto, presence)
			err := dispatcher.BroadcastMessage(OpCodeJua, player.CreateJua(true, juaCard.id), sendto, nil, true)
			if err != nil {
				logger.Info("cause an error when broadcasting", err)
			}
		} else {
			sendto = append(sendto, presence)
			err := dispatcher.BroadcastMessage(OpCodeJua, player.CreateJua(false, 53), sendto, nil, true)
			if err != nil {
				logger.Info("cause an error when broadcasting", err)
			}
		}
	}
}
func (room *PokdengRoom) saveResult(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) {
	metadata := map[string]interface{}{
		"game_result": "game_result",
	}
	for _, player := range room.game.players {
		chipGainsAfterTax := room.chipGainAfterTax(player, room.game.isDealerBot())
		player.ChipGain = chipGainsAfterTax

		content := map[string]int64{
			"gold": player.ChipGain, // Add 1000 coins to the user's wallet.
		}
		if _, _, err := nk.WalletUpdate(ctx, player.UID, content, metadata, true); err != nil {
			logger.Error("User wallet update error: %v", err.Error())
		}
	}

	if !room.game.isDealerBot() {
		chipGainsAfterTax := room.chipGainAfterTax(room.game.dealer, false)
		room.game.dealer.ChipGain = chipGainsAfterTax
		content := map[string]int64{
			"gold": room.game.dealer.ChipGain, // Add 1000 coins to the user's wallet.
		}
		if _, _, err := nk.WalletUpdate(ctx, room.game.dealer.UID, content, metadata, true); err != nil {
			logger.Error("User wallet update error: %v", err.Error())
		}
	}
}

func (room *PokdengRoom) saveResultJub(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) {
	metadata := map[string]interface{}{
		"game_result": "game_result",
	}

	if !room.game.isDealerBot() {

		chipGainsAfterTax := room.chipGainAfterTax(room.game.dealer, false)
		room.game.dealer.ChipGain = chipGainsAfterTax
		content := map[string]int64{
			"gold": room.game.dealer.ChipGain, // Add 1000 coins to the user's wallet.
		}
		if _, _, err := nk.WalletUpdate(ctx, room.game.dealer.UID, content, metadata, true); err != nil {
			logger.Error("User wallet update error: %v", err.Error())
		}
	}
}

func (room *PokdengRoom) savePlayerMoney(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, player *PokdengPlayer) {
	metadata := map[string]interface{}{
		"game_result": "game_result",
	}
	content := map[string]int64{
		"gold": player.ChipGain, // Add 1000 coins to the user's wallet.
	}
	if _, _, err := nk.WalletUpdate(ctx, player.UID, content, metadata, true); err != nil {
		logger.Error("User wallet update error: %v", err.Error())
	}
}

func (room *PokdengRoom) chipGainAfterTax(player *PokdengPlayer, isdealerBot bool) int64 {
	if player.ChipGain < 0 {
		return player.ChipGain
	}

	taxRate := room.tax

	if isdealerBot {
		// Normal 10% Tax.
		taxRate = 5
	}
	tax := ((float64(player.ChipGain) * taxRate) / 100)
	go handler.SaveTax(int(math.Floor(tax)))
	chipGainAfterTax := int64(math.Floor(float64(player.ChipGain) - tax))
	return chipGainAfterTax
}

func (room *PokdengRoom) GetAccountWallet(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, id string) float64 {
	account, err := nk.AccountGetId(ctx, id)
	var gold map[string]interface{}

	if err != nil {
		// Handle error.
	} else {
		logger.Info("Wallet is: %v", account.Wallet)
		err := json.Unmarshal([]byte(account.Wallet), &gold)
		if err == nil {
			return gold["gold"].(float64)
		} else {
			return 0
		}
	}
	return 0
}

func (room *PokdengRoom) verifyDealerChip(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, id string, maxBetRate uint64) bool {
	gold := room.GetAccountWallet(ctx, logger, nk, id)
	if gold >= float64(maxBetRate)*20 {
		return true
	}
	return false
}
