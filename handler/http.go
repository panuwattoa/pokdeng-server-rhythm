package handler

type LogPlay struct {
	UID        string
	Name       string
	Bet        int
	MyCard     string
	DealerCard string
	IsDealer   bool
	ChipGain   int
	DealerUID  string
}

type LogDealerPlay struct {
	BetRate int
	Value   int
}

type Payment struct {
	UID     string
	Request string
	Time    string
	IsPass  bool
}

func SaveLogDealer(p LogDealerPlay) {
	// bytesRepresentation, _ := json.Marshal(p)
	// http.Post("http://10.128.0.2:3111/log_dealer_pay", "application/json", bytes.NewBuffer(bytesRepresentation))
}

func SaveLogPlay(p LogPlay) {
	// bytesRepresentation, _ := json.Marshal(p)
	// http.Post("http://10.128.0.2:3111/log_pay", "application/json", bytes.NewBuffer(bytesRepresentation))
}

func SaveTax(value int) {
	// http.Get("http://10.128.0.2:3111/tax/" + strconv.Itoa(value))
}
