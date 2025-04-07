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
