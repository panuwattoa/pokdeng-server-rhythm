package systems

type IAPProduct struct {
	IAP []Product `json:"iap"`
}
type Product struct {
	ProductID       string `json:"product_id"`
	ProductNameText string `json:"product_name"`
	Gold            uint   `json:"gold"`
	Diamond         uint   `json:"diamond"`
	Bonus           uint   `json:"bonus"`
}

type VedioAds struct {
	Number    uint `json:"number"`
	ResetTime uint `json:"resettime"`
}

type UserVideoAds struct {
	NumWatch int64  `json:"num_watch"`
	Date     string `json:"Date"`
}

type UserData struct {
	NumSpecialIAP uint `json:"num_special_iap"`
}
