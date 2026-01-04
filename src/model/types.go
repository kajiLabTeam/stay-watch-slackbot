package model

func (t *Type) Create() error {
	if err := db.Create(t).Error; err != nil {
		return err
	}
	return nil
}

func (t *Type) ReadByID() error {
	if err := db.First(t, t.ID).Error; err != nil {
		return err
	}
	return nil
}

func (t *Type) ReadByName() error {
	if err := db.Where("name = ?", t.Name).First(t).Error; err != nil {
		return err
	}
	return nil
}

func (t *Type) ReadAll() ([]Type, error) {
	var types []Type
	if err := db.Find(&types).Error; err != nil {
		return types, err
	}
	return types, nil
}
