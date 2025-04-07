package model

func (c *Correspond) Create() (err error) {
	if err := db.Create(&c).Error; err != nil {
		return err
	}
	return nil
}

func (c *Correspond) ReadByID() error {
	if err := db.First(&c, c.ID).Error; err != nil {
		return err
	}
	return nil
}

func (c *Correspond) ReadByTagID() (corresponds []Correspond, err error) {
	if err := db.Where("tag_id = ?", c.TagID).Find(&corresponds).Error; err != nil {
		return corresponds, err
	}
	return corresponds, nil
}

func (c *Correspond) ReadByUserID() (corresponds []Correspond, err error) {
	if err := db.Where("user_id = ?", c.UserID).Find(&corresponds).Error; err != nil {
		return corresponds, err
	}
	return corresponds, nil
}
