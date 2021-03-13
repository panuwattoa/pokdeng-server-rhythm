package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"pokdeng-server/game"
	"pokdeng-server/handler"
	"sort"
	"strconv"

	"github.com/heroiclabs/nakama-common/runtime"
)

const (
	pokdeng     = "pokeng-game"
	GameVersion = 4
)

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	logger.Info("module loaded")
	createRoomMatch := func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
		return &game.PokdengRoom{}, nil
	}

	if err := initializer.RegisterMatch(pokdeng, createRoomMatch); err != nil {
		return err
	}

	if err := initializer.RegisterRpc("get_matches", GetRooms); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterRpc("check_version", CheckVersion); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterRpc("check_money_list", RequestMoneyList); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterRpc("save_bank", RequestSaveBank); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterRpc("request_payment", RequestPayment); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterRpc("request_paymentv2", RequestPaymentV2); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	if err := initializer.RegisterRpc("save_bankV2", RequestSaveBankV2); err != nil {
		logger.Error("Unable to register: %v", err)
		return err
	}

	return nil
}

// get match list
func GetRooms(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var input map[string]interface{}
	err := json.Unmarshal([]byte(payload), &input)
	if err != nil {
		return "", err
	}

	minSize := 0
	maxSize := 7
	query := input["gold"].(string)
	if users, err := nk.AccountsGetId(ctx, []string{
		input["uid"].(string),
	}); err == nil {
		if len(users) == 0 {
			return "", errors.New("Not Found user")
		}
		for _, u := range users {
			var gold map[string]interface{}
			wallet := u.GetWallet()
			err := json.Unmarshal([]byte(wallet), &gold)
			if err != nil {
				return "", err
			}
			i, err := strconv.Atoi(query)
			if err != nil {
				return "", err
			}
			if gold["gold"].(float64) < float64(i*5) {
				return "", errors.New("Not enoung money")
			}
		}
	} else {
		return "", err
	}
	if matches, err := nk.MatchList(ctx, 100, true, query, &minSize, &maxSize, ""); err != nil {
		return "", err
	} else {
		if len(matches) > 0 {
			sort.Slice(matches, func(i, j int) bool {
				return matches[i].Size < matches[j].Size
			})
			if matches[0].Size < 7 {
				return matches[0].MatchId, nil
			}
		}
	}

	if matchId, err := nk.MatchCreate(ctx, pokdeng, input); err != nil {
		return "", err
	} else {
		return matchId, nil
	}
}

func CheckVersion(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var input map[string]interface{}
	err := json.Unmarshal([]byte(payload), &input)
	if err != nil {
		return "false", err
	}
	query := input["version"].(float64)
	if int(query) < GameVersion {
		return "false", nil
	}
	return "true", nil
}

func RequestPayment(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var input map[string]interface{}
	err := json.Unmarshal([]byte(payload), &input)
	if err != nil {
		return "false", err
	}
	return "true", nil
}

func RequestPaymentV2(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var input map[string]interface{}
	err := json.Unmarshal([]byte(payload), &input)
	if err != nil {
		return "false", err
	}
	id := input["id"].(string)
	money := input["money"].(string)
	time := input["time"].(string)
	result, err := handler.RequestPayment(ctx, id, money, time)
	logger.Debug("RequestPaymentV2 %v", err)
	if err != nil {
		return "false", errors.New("Not pass")
	}
	if result {
		f, err := strconv.ParseFloat(money, 32)
		if err != nil {
			logger.Debug("err ParseFloat %v", err)
			return "false", errors.New("Not pass")
		}

		content := map[string]int64{
			"gold": int64(math.Round(f)), // Add 1000 coins to the user's wallet.
		}
		metadata := map[string]interface{}{
			"payment": "+" + money,
		}
		if _, _, err := nk.WalletUpdate(ctx, id, content, metadata, true); err != nil {
			logger.Debug("User wallet update error: %v", err.Error())
			return "false", errors.New("Not pass")
		}
		return "true", nil
	} else {
		return "false", nil
	}
}

func RequestMoneyList(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var input map[string]interface{}
	err := json.Unmarshal([]byte(payload), &input)
	if err != nil {
		return "false", err
	}
	return "{\"money\": [100,200,300,400,500,1000]}", nil
}

func RequestSaveBank(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var input map[string]interface{}
	err := json.Unmarshal([]byte(payload), &input)
	if err != nil {
		return "false", err
	}
	return "true", nil
}

func RequestSaveBankV2(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var input map[string]interface{}
	err := json.Unmarshal([]byte(payload), &input)
	if err != nil {
		return "false", err
	}
	email := input["email"].(string)
	bank_no := input["bank_no"].(string)
	bank_name := input["bank_name"].(string)
	if err = handler.SaveBank(email, bank_no, bank_name); err != nil {
		logger.Error("User SaveBank  error: %v", err.Error())
		logger.Debug("User SaveBank  error: %v", err.Error())
	}
	return "true", nil
}
