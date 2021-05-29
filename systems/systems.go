package systems

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math/rand"
	"pokdeng-server/config"
	"sort"
	"strconv"
	"time"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

const RFC3339FullDate = "2006-01-02"

var IAPProductList map[string]Product
var VedioAdsCong VedioAds
var IAPRaw IAPProduct

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

		// write
		userInit := UserData{
			NumSpecialIAP: 1,
		}
		b, _ := json.Marshal(userInit)
		objectsW := []*runtime.StorageWrite{
			{
				Collection:      "user",
				Key:             "data",
				UserID:          userID,
				Value:           string(b),
				PermissionRead:  1,
				PermissionWrite: 1,
			},
		}
		if _, err := nk.StorageWrite(ctx, objectsW); err != nil {
			// Handle error.
			logger.Error("User wallet StorageWrite: %v", err.Error())
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
		// write
		userInit := UserData{
			NumSpecialIAP: 1,
		}
		b, _ := json.Marshal(userInit)
		objectsW := []*runtime.StorageWrite{
			{
				Collection:      "user",
				Key:             "data",
				UserID:          userID,
				Value:           string(b),
				PermissionRead:  1,
				PermissionWrite: 1,
			},
		}
		if _, err := nk.StorageWrite(ctx, objectsW); err != nil {
			// Handle error.
			logger.Error("User wallet StorageWrite: %v", err.Error())
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
func RequestPaymentGoogle(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userId, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		// User ID not found in the context.
		return "false", errors.New("can't find user id")
	}

	logger.Debug(" %v", payload)

	purchase := &api.ValidatePurchaseResponse{}
	purchase, err := nk.PurchaseValidateGoogle(ctx, userId, payload)
	if err != nil {
		logger.Error("got PurchaseValidateGoogle %v", err)
		return "ไม่สำเร็จ", errors.New("can't validate payload")
	}

	if purchase != nil {
		var isPass bool
		var numGold int64
		for _, v := range purchase.ValidatedPurchases {
			if product, ok := IAPProductList[v.ProductId]; ok {
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

func RequestPaymentApple(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userId, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		// User ID not found in the context.
		return "false", errors.New("can't find user id")
	}

	logger.Debug(" %v", payload)

	purchase := &api.ValidatePurchaseResponse{}

	purchase, err := nk.PurchaseValidateApple(ctx, userId, payload)
	if err != nil {
		logger.Error("got PurchaseValidateApple %v", err)
		return "ไม่สำเร็จ", errors.New("can't validate payload")
	}

	if purchase != nil {
		var isPass bool
		var numGold int64
		for _, v := range purchase.ValidatedPurchases {
			if product, ok := IAPProductList[v.ProductId]; ok {
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
	objectIds := []*runtime.StorageRead{
		{
			Collection: "user_video_ads",
			Key:        "data",
			UserID:     userId,
		},
	}
	uAds := UserVideoAds{}
	objects, err := nk.StorageRead(ctx, objectIds)
	if err != nil {
		logger.Error("User StorageRead: %v", err.Error())
		// Handle error.
		return "ผิดพลาด", nil
	} else {
		for _, object := range objects {
			if object.Key == "data" {
				if err := json.Unmarshal([]byte(object.Value), &uAds); err != nil {
					logger.Error("Unable to read user_video_ads Unmarshal: %v", err)
					return "ผิดพลาด", nil
				}
				if uint(uAds.NumWatch) >= VedioAdsCong.Number {
					t, err := time.Parse(RFC3339FullDate, uAds.Date)
					if err != nil {
						return "ผิดพลาด", nil
					}
					if DateEqual(t, time.Now()) {
						return "นายท่านดู ads ครบจำนวนแล้ว\nสามารถดูได้อีกวันถัดไป", nil
					}
					uAds.NumWatch = 0
				}
			}
		}
	}

	chip := int64(500)
	randomReward100 := 30
	randomReward200 := 40
	randomReward300 := 60
	randomReward400 := 95
	randomReward500 := 100
	rate := rand.Intn(100) + 1
	if rate <= randomReward100 {
		chip = 500
	} else if rate <= randomReward200 {
		chip = 1000
	} else if rate <= randomReward300 {
		chip = 1500
	} else if rate <= randomReward400 {
		chip = 2500
	} else if rate <= randomReward500 {
		chip = 3000
	}
	content := map[string]int64{
		"gold": chip,
	}
	metadata := map[string]interface{}{
		"random": rate,
	}

	uAds.Date = time.Now().Format(RFC3339FullDate)
	uAds.NumWatch = uAds.NumWatch + 1
	b, err := json.Marshal(uAds)

	if err != nil {
		logger.Error("User Marshal RequestClaimVideoReward error: %v", err.Error())
		return "ผิดพลาด", nil
	}
	// write
	objectsW := []*runtime.StorageWrite{
		{
			Collection:      "user_video_ads",
			Key:             "data",
			UserID:          userId,
			Value:           string(b),
			PermissionRead:  1,
			PermissionWrite: 1,
		},
	}
	if _, err := nk.StorageWrite(ctx, objectsW); err != nil {
		// Handle error.
		logger.Error("User wallet StorageWrite: %v", err.Error())
	}

	if _, _, err := nk.WalletUpdate(ctx, userId, content, metadata, true); err != nil {
		logger.Error("User wallet update error: %v", err.Error())
		return "ผิดพลาด", nil
	}
	return "สุ่มได้\n " + strconv.Itoa(int(chip)) + " ทอง", nil
}

func CheckCanWatchVideoAds(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userId, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		// User ID not found in the context.
		return "ผิดพลาด", errors.New("can't find user id")
	}

	objectIds := []*runtime.StorageRead{
		{
			Collection: "user_video_ads",
			Key:        "data",
			UserID:     userId,
		},
	}
	objects, err := nk.StorageRead(ctx, objectIds)
	if err != nil {
		logger.Error("User StorageRead: %v", err.Error())
		// Handle error.
		return "ผิดพลาด", errors.New("StorageRead")
	} else {
		for _, object := range objects {
			if object.Key == "data" {
				uAds := UserVideoAds{}
				if err := json.Unmarshal([]byte(object.Value), &uAds); err != nil {
					logger.Error("Unable to read user_video_ads Unmarshal: %v", err)
					return "ผิดพลาด", nil
				}
				if uint(uAds.NumWatch) >= VedioAdsCong.Number {
					t, err := time.Parse(RFC3339FullDate, uAds.Date)
					if err != nil {
						return "ผิดพลาด", nil
					}
					if DateEqual(t, time.Now()) {
						return "นายท่านดู ads ครบจำนวนแล้ว\nสามารถดูได้อีกวันถัดไป", nil
					}
				}
			}
		}
	}
	return "true", nil
}

func CheckUserCanBuySpecialIAP(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userId, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		// User ID not found in the context.
		return "ผิดพลาด", errors.New("can't find user id")
	}

	objectIds := []*runtime.StorageRead{
		{
			Collection: "user",
			Key:        "data",
			UserID:     userId,
		},
	}
	objects, err := nk.StorageRead(ctx, objectIds)
	if err != nil {
		logger.Error("User StorageRead: %v", err.Error())
		// Handle error.
		return "ผิดพลาด", errors.New("StorageRead")
	} else {
		for _, object := range objects {
			if object.Key == "data" {
				u := UserData{}
				if err := json.Unmarshal([]byte(object.Value), &u); err != nil {
					logger.Error("Unable to read user_video_ads Unmarshal: %v", err)
					return "ผิดพลาด", nil
				}
				if u.NumSpecialIAP >= config.NumUserCanBuySpecialIAP {
					return "false", nil
				}
			}
		}
	}
	return "true", nil
}

func GetIAPList(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	b, err := json.Marshal(IAPRaw)
	if err != nil {
		logger.Error(" Marshal: %v", err.Error())
		// Handle error.
		return "ผิดพลาด", errors.New("Marshal")
	}
	return string(b), nil
}

func BuySpecial(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userId, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		// User ID not found in the context.
		return "ผิดพลาด", errors.New("can't find user id")
	}
	objectIds := []*runtime.StorageRead{
		{
			Collection: "user",
			Key:        "data",
			UserID:     userId,
		},
	}
	objects, _ := nk.StorageRead(ctx, objectIds)
	for _, object := range objects {
		if object.Key == "data" {
			u := UserData{}
			if err := json.Unmarshal([]byte(object.Value), &u); err != nil {
				logger.Error("Unable to read user_video_ads Unmarshal: %v", err)
			}
			u.NumSpecialIAP += 1
			b, _ := json.Marshal(u)
			objectsW := []*runtime.StorageWrite{
				{
					Collection:      "user",
					Key:             "data",
					UserID:          userId,
					Value:           string(b),
					PermissionRead:  1,
					PermissionWrite: 1,
				},
			}
			if _, err := nk.StorageWrite(ctx, objectsW); err != nil {
				// Handle error.
				logger.Error("User wallet StorageWrite: %v", err.Error())
			}
		}
	}
	return "", nil
}
func DateEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
