package service

import (
	"errors"
	"log"

	"github.com/kajiLabTeam/stay-watch-slackbot/model"
)

func RegisterUser(slackUserID string, userName string) (model.User, error) {
	// userNameをもとに滞在ウォッチからユーザ情報を取得
	users, err := GetStayWatchMember()
	if err != nil {
		return model.User{}, err
	}
	user := model.User{
		Name:        userName,
		SlackID:     slackUserID,
		StayWatchID: int64(0),
	}
	if err := user.ReadByName(); err != nil {
		return user, err
	}
	if user.StayWatchID != 0 {
		err := errors.New("user already exists")
		return user, err
	}
	for i, u := range users {
		if u.Name == userName {
			user.StayWatchID = u.ID
			break
		}
		if i == len(users)-1 {
			err := errors.New("user not found")
			return user, err
		}
	}
	if err := user.Create(); err != nil {
		return user, err
	}

	// アイコンURLをSlackから取得してDBに保存（失敗しても登録自体は成功とする）
	if iconURL, err := fetchSlackIconURL(slackUserID); err == nil {
		user.IconURL = iconURL
		if err := user.UpdateIconURL(); err != nil {
			log.Printf("failed to save icon_url for user %s: %v", slackUserID, err)
		}
	} else {
		log.Printf("failed to fetch icon URL for user %s: %v", slackUserID, err)
	}

	return user, nil
}

// fetchSlackIconURL は Slack API からユーザのアイコン画像 URL を取得する
func fetchSlackIconURL(slackUserID string) (string, error) {
	slackUser, err := slackClient.GetUserInfo(slackUserID)
	if err != nil {
		return "", err
	}
	return slackUser.Profile.Image192, nil
}

// ListAllUsers は登録済みユーザの一覧を取得する
func ListAllUsers() ([]model.User, error) {
	u := model.User{}
	return u.ReadAll()
}

// DeleteUserByName は指定した名前のユーザを削除する
func DeleteUserByName(name string) error {
	user := model.User{Name: name}
	if err := user.ReadByName(); err != nil {
		return err
	}
	if user.ID == 0 {
		return errors.New("user not found")
	}
	return user.Delete()
}

// DeleteOBUsers はStayWatch側でOBタグ（id:13, name:"OB"）が付与されている
// ユーザーを一括削除し、削除したユーザー名の一覧を返す
func DeleteOBUsers() ([]string, error) {
	u := model.User{}
	users, err := u.ReadAll()
	if err != nil {
		return nil, err
	}

	var deleted []string
	for i := range users {
		detail, err := GetStayWatchUserDetail(users[i].StayWatchID)
		if err != nil {
			log.Printf("failed to fetch StayWatch detail for user %s: %v", users[i].Name, err)
			continue
		}
		if !hasOBTag(detail) {
			continue
		}
		if err := users[i].Delete(); err != nil {
			log.Printf("failed to delete OB user %s: %v", users[i].Name, err)
			continue
		}
		deleted = append(deleted, users[i].Name)
	}
	return deleted, nil
}

// RefreshAllUserIcons は全ユーザのアイコン URL を Slack から取得し直してDBを更新する
func RefreshAllUserIcons() (int, error) {
	u := model.User{}
	users, err := u.ReadAll()
	if err != nil {
		return 0, err
	}

	updated := 0
	for i := range users {
		iconURL, err := fetchSlackIconURL(users[i].SlackID)
		if err != nil {
			log.Printf("failed to fetch icon URL for user %s: %v", users[i].SlackID, err)
			continue
		}
		users[i].IconURL = iconURL
		if err := users[i].UpdateIconURL(); err != nil {
			log.Printf("failed to update icon_url for user %s: %v", users[i].SlackID, err)
			continue
		}
		updated++
	}
	return updated, nil
}
