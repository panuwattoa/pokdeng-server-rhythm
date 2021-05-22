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

	if err := initializer.RegisterRpc("request_payment", systems.RequestPayment); err != nil {
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

	if err := initializer.RegisterAfterAuthenticateDevice(systems.InitializeUser); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterAfterAuthenticateFacebook(systems.InitializeFacebookUser); err != nil {
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
				var iapProduct systems.IAPProduct
				if err := json.Unmarshal([]byte(object.Value), &iapProduct); err != nil {
					logger.Error("Unable to read IAPStorageKey: %v", err)
					continue
				}
				systems.IAPProductList = make(map[string]systems.Product)
				for _, v := range iapProduct.IAP {
					systems.IAPProductList[v.ProductID] = v
				}
			}
			if object.Key == config.VideoAdsKey {
				if err := json.Unmarshal([]byte(object.Value), &systems.VedioAdsCong); err != nil {
					logger.Error("Unable to read VedioAdsCong: %v", err)
					continue
				}
			}

		}
	}
	return nil
}
