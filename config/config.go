package config

import "github.com/heroiclabs/nakama-common/runtime"

const (
	Pokdeng                 = "pokeng-game"
	GameVersion             = 4
	SettingStorageKey       = "setting"
	IAPStorageKey           = "iap-product"
	VideoAdsKey             = "videoads"
	NumUserCanBuySpecialIAP = 1
)

var InitStorage = []*runtime.StorageRead{
	{
		Collection: "configuration",
		Key:        IAPStorageKey,
	},
	{
		Collection: "configuration",
		Key:        VideoAdsKey,
	},
}
