package model

func (t *Tag) Create() (err error) {
	if err := db.Create(t).Error; err != nil {
		return err
	}
	return nil
}

func (t *Tag) ReadByID() error {
	if err := db.Limit(1).Find(t, t.ID).Error; err != nil {
		return err
	}
	return nil
}

func (t *Tag) ReadByName() error {
	if err := db.Where("name = ?", t.Name).Limit(1).Find(t).Error; err != nil {
		return err
	}
	return nil
}

func (t *Tag) ReadAll() ([]Tag, error) {
	var tags []Tag
	if err := db.Find(&tags).Error; err != nil {
		return tags, err
	}
	return tags, nil
}
