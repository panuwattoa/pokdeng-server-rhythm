package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"math/rand"
	"pokdeng-server/config"
	"pokdeng-server/game"
	"pokdeng-server/systems"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
)

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	logger.Info("module loaded")
	rand.Seed(time.Now().UnixNano())

	createRoomMatch := func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
		return &game.PokdengRoom{}, nil
	}

	if err := initializer.RegisterMatch(config.Pokdeng, createRoomMatch); err != nil {
		return err
	}

	if err := initializer.RegisterRpc("get_matches", systems.GetRooms); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterRpc("check_version", systems.CheckVersion); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterRpc("request_payment_goolge", systems.RequestPaymentGoogle); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterRpc("request_payment_apple", systems.RequestPaymentApple); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterRpc("request_claim_video_reward", systems.RequestClaimVideoReward); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterRpc("request_check_video_reward", systems.CheckCanWatchVideoAds); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterRpc("request_iap_list", systems.GetIAPList); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}
	if err := initializer.RegisterRpc("buy_special", systems.BuySpecial); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterRpc("request_check_special_iap", systems.CheckUserCanBuySpecialIAP); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterAfterAuthenticateDevice(systems.InitializeUser); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterAfterAuthenticateFacebook(systems.InitializeFacebookUser); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterAfterAuthenticateApple(systems.InitializeApplekUser); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterRpc("request_check_can_buy_special_iap", systems.CheckUserCanBuySpecialIAP); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterRpc("request_check_user_data", systems.CheckUserData); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterRpc("request_claim_play_reward", systems.ClaimPlayReward); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterRpc("request_claim_daily_reward", systems.ClaimUserDailyReward); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterRpc("request_login_data", systems.LoginRequestData); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	// init config

	objects, err := nk.StorageRead(ctx, config.InitStorage)
	if err != nil {
		// Handle error.
		logger.Error("Unable to read storage: %v", err)
		return err
	} else {
		for _, object := range objects {
			logger.Info("value: %s", object.Value)
			if object.Key == config.IAPStorageKey {
				if err := json.Unmarshal([]byte(object.Value), &systems.IAPRaw); err != nil {
					logger.Error("Unable to read IAPStorageKey: %v", err)
					continue
				}
				systems.IAPProductList = make(map[string]systems.Product)
				for _, v := range systems.IAPRaw.IAP {
					systems.IAPProductList[v.ProductID] = v
				}
			}
			if object.Key == config.VideoAdsKey {
				if err := json.Unmarshal([]byte(object.Value), &systems.VedioAdsCong); err != nil {
					logger.Error("Unable to read VedioAdsCong: %v", err)
					continue
				}
			}

			if object.Key == config.DailyRewardKey {
				logger.Debug("Load DailyRewardKey... ")
				if err := json.Unmarshal([]byte(object.Value), &systems.DailyRewardList); err != nil {
					logger.Error("Unable to read DailyRewardKey: %v", err)
					continue
				}
			}

			if object.Key == config.PlayRewardKey {
				logger.Debug("Load PlayRewardKey... ")
				if err := json.Unmarshal([]byte(object.Value), &systems.PlayRewardList); err != nil {
					logger.Error("Unable to read PlayRewardKey: %v", err)
					continue
				}
			}

		}
	}
	return nil
}
