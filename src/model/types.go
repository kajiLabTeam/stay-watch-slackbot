package model

func (t *Types) Create() error {
	if err := db.Create(t).Error; err != nil {
		return err
	}
	return nil
}

func (t *Types) ReadByID() error {
	if err := db.First(t, t.ID).Error; err != nil {
		return err
	}
	return nil
}

func (t *Types) ReadByName() error {
	if err := db.Where("name = ?", t.Name).First(t).Error; err != nil {
		return err
	}
	return nil
}

func (t *Types) ReadAll() ([]Types, error) {
	var types []Types
	if err := db.Find(&types).Error; err != nil {
		return types, err
	}
	return types, nil
}
