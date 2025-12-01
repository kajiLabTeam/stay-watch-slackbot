package model

func (c *Correspondence) Create() error {
	if err := db.Create(c).Error; err != nil {
		return err
	}
	return nil
}

func (c *Correspondence) ReadByID() error {
	if err := db.Preload("Event").Preload("Tag").First(c, c.ID).Error; err != nil {
		return err
	}
	return nil
}

func (c *Correspondence) ReadByEventID() ([]Correspondence, error) {
	var correspondences []Correspondence
	if err := db.Preload("Event").Preload("Tag").Where("event_id = ?", c.EventID).Find(&correspondences).Error; err != nil {
		return correspondences, err
	}
	return correspondences, nil
}

func (c *Correspondence) ReadByTagID() ([]Correspondence, error) {
	var correspondences []Correspondence
	if err := db.Preload("Event").Preload("Tag").Where("tag_id = ?", c.TagID).Find(&correspondences).Error; err != nil {
		return correspondences, err
	}
	return correspondences, nil
}
