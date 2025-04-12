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

func (u *User) ReadAll() ([]User, error) {
	var users []User
	if err := db.Preload("Corresponds").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
