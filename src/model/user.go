package model

func (u *User) Create() (err error) {
	if err := db.Create(u).Error; err != nil {
		return err
	}
	return nil
}

func (u *User) ReadByID() error {
	if err := db.First(u, u.ID).Error; err != nil {
		return err
	}
	return nil
}

func (u *User) ReadBySlackID() error {
	if err := db.Where("slack_id = ?", u.SlackID).Limit(1).Find(u).Error; err != nil {
		return err
	}
	return nil
}

func (u *User) ReadByName() error {
	if err := db.Where("name = ?", u.Name).Limit(1).Find(u).Error; err != nil {
		return err
	}
	return nil
}

func (u *User) ReadByStayWatchID() error {
	if err := db.Where("stay_watch_id = ?", u.StayWatchID).Limit(1).Find(u).Error; err != nil {
		return err
	}
	return nil
}

// ReadByStayWatchIDs は複数の StayWatchID からユーザーをバッチで取得する
// filterByThreshold などで N+1 クエリ問題を回避するために使用
func (u *User) ReadByStayWatchIDs(ids []int64) ([]User, error) {
	var users []User
	if err := db.Where("stay_watch_id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (u *User) ReadAll() ([]User, error) {
	var users []User
	if err := db.Preload("Corresponds").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
