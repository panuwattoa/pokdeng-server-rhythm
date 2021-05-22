package systems

type IAPProduct struct {
	IAP []Product `json:"iap"`
}
type Product struct {
	ProductID string `json:"product_id"`
	Gold      uint   `json:"gold"`
	Diamond   uint   `json:"diamond"`
	Bonus     uint   `json:"bonus"`
}

type VedioAds struct {
	Number    uint `json:"videoads"`
	ResetTime uint `json:"resettime"`
}
