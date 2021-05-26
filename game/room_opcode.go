package game

const (
	OpCodeTerminate   = iota // Server -> client when the party is being terminated by the server.
	OpCodeJoinRequest        // Server -> client when a user requests to join the party.
	OpCodeOtherJoin
	OpCodeLeave
	OpCodeStanUp
	OpCodeFightForDealer
	OpCodeRequestDealer
	OpCodeUserBet
	OpCodeUserPok
	OpCodeJua
	OpCodeJub
	OpCodeSetDealer
	OpCodePok
	OpCodeChnageTurn
	OpCodeBotJua
	OpCodeDealCard
	OpCodeGameStartBet
	OpCodeOpeningTable
	OpCodeNotJua
	OpCodeClearTable
	OpBotDealerCard
	OpCodeGameResult
	OpCodeSelfLeave
	OpCodeOpenRequestDealer
	OpCodeDealerResult
	OpCodeCancelDealer
)

const (
	GameOpCodeNone = iota
	GameOpCodeWaitOpenTable
	GameOpCodeDealCard
	GameOpCodePlaying
	GameStartBet
	GameOpCodeWaitingBet
	GameProcessScore
	GameChangeTurn
	GameOpCodeDealerPlay
	GameClearTable
	GamePlayerReadyPlay
	GameInitGamePlay
	GameProcessScoreWithJub
)
