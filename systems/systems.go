package systems

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"pokdeng-server/config"
	"sort"
	"strconv"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

var IAPProductList map[string]Product
var VedioAdsCong VedioAds

func InitializeUser(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, out *api.Session, in *api.AuthenticateDeviceRequest) error {
	if out.Created {
		// Only run this logic if the account that has authenticated is new.
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return errors.New("Invalid context")
		}
		changeset := map[string]int64{
			"gold": 10000, // Add 10 coins to the user's wallet.
		}
		metadata := map[string]interface{}{
			"newuser": 10000,
		}
		if _, _, err := nk.WalletUpdate(ctx, userID, changeset, metadata, true); err != nil {
			// Handle error.
			logger.Error("Unable to WalletUpdate new user : %v", err)
		}
	}
	return nil
}

func InitializeFacebookUser(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, out *api.Session, in *api.AuthenticateFacebookRequest) error {
	if out.Created {
		// Only run this logic if the account that has authenticated is new.
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return errors.New("Invalid context")
		}
		changeset := map[string]int64{
			"gold": 10000, // Add 10 coins to the user's wallet.
		}
		metadata := map[string]interface{}{
			"newuser": 10000,
		}
		if _, _, err := nk.WalletUpdate(ctx, userID, changeset, metadata, true); err != nil {
			// Handle error.
			logger.Error("Unable to WalletUpdate new user : %v", err)
		}
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

	if matchId, err := nk.MatchCreate(ctx, config.Pokdeng, input); err != nil {
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
	if int(query) < config.GameVersion {
		return "false", nil
	}
	return "true", nil
}

func RequestPayment(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userId, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		// User ID not found in the context.
		return "false", errors.New("can't find user id")
	}

	var input map[string]interface{}
	err := json.Unmarshal([]byte(payload), &input)
	if err != nil {
		logger.Error("got Unmarshal err %v", err)
		return "ไม่สำเร็จ", err
	}
	receipt, ok := input["receipt"].(string)
	if !ok {
		logger.Error("got receipt err %v", err)
		return "ไม่สำเร็จ", errors.New("can't find receipt")
	}

	platfrom, ok := input["platfrom"].(string)
	if !ok {
		logger.Error("got platfrom err %v", err)
		return "ไม่สำเร็จ", errors.New("can't find platfrom")
	}
	purchase := &api.ValidatePurchaseResponse{}

	if platfrom == "apple" {
		purchase, err = nk.PurchaseValidateApple(ctx, userId, receipt)
		if err != nil {
			logger.Error("got PurchaseValidateApple %v", err)
			return "ไม่สำเร็จ", errors.New("can't validate payload")
		}
	} else if platfrom == "google" {
		purchase, err = nk.PurchaseValidateGoogle(ctx, userId, receipt)
		if err != nil {
			logger.Error("got PurchaseValidateGoogle %v", err)
			return "ไม่สำเร็จ", errors.New("can't validate payload")
		}
	}

	if purchase != nil {
		var isPass bool
		var numGold int64
		for _, v := range purchase.ValidatedPurchases {
			if product, ok := IAPProductList[v.ProductId]; ok {
				logger.Info("got product %v", product)
				isPass = true
				gold := int64(product.Gold) + int64(product.Bonus)
				numGold += gold
				content := map[string]int64{
					"gold": gold,
				}
				metadata := map[string]interface{}{
					"iap":           product.ProductID,
					"time":          v.PurchaseTime,
					"store":         v.Store,
					"TransactionId": v.TransactionId,
					"env":           v.Environment,
					"bonus":         product.Bonus,
				}
				if _, _, err := nk.WalletUpdate(ctx, userId, content, metadata, true); err != nil {
					logger.Error("User wallet update error: %v", err.Error())
				}
			}
		}
		if isPass {
			return "นายท่านได้รับ\n" + strconv.Itoa(int(numGold)) + " ทอง", nil
		}
	}
	return "ไม่สำเร็จ", errors.New("can't validate payload")
}

func RequestClaimVideoReward(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userId, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		// User ID not found in the context.
		return "ผิดพลาด", errors.New("can't find user id")
	}
	content := map[string]int64{
		"gold": 100,
	}
	metadata := map[string]interface{}{}
	if _, _, err := nk.WalletUpdate(ctx, userId, content, metadata, true); err != nil {
		logger.Error("User wallet update error: %v", err.Error())
	}
	return "true", nil
}

func CheckCanWatchVideoAds(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	_, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		// User ID not found in the context.
		return "ผิดพลาด", errors.New("can't find user id")
	}
	return "true", nil
}
