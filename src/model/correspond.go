package model

func (c *Correspond) Create() error {
	if err := db.Create(c).Error; err != nil {
		return err
	}
	return nil
}

func (c *Correspond) ReadByID() error {
	if err := db.Preload("Event").Preload("User").First(c, c.ID).Error; err != nil {
		return err
	}
	return nil
}

func (c *Correspond) ReadByEventID() ([]Correspond, error) {
	var corresponds []Correspond
	if err := db.Preload("Event").Preload("User").Where("event_id = ?", c.EventID).Find(&corresponds).Error; err != nil {
		return corresponds, err
	}
	return corresponds, nil
}

func (c *Correspond) ReadByUserID() ([]Correspond, error) {
	var corresponds []Correspond
	if err := db.Preload("Event").Preload("User").Where("user_id = ?", c.UserID).Find(&corresponds).Error; err != nil {
		return corresponds, err
	}
	return corresponds, nil
}
