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

// ReadByUserIDs は複数の UserID から Correspond をバッチで取得する
// GroupByEvent などで N+1 クエリ問題を回避するために使用
func (c *Correspond) ReadByUserIDs(userIDs []uint) ([]Correspond, error) {
	var corresponds []Correspond
	if err := db.Preload("Event").Where("user_id IN ?", userIDs).Find(&corresponds).Error; err != nil {
		return nil, err
	}
	return corresponds, nil
}
