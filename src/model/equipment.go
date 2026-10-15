package model

func (e *Equipment) Create() error {
	if err := db.Create(e).Error; err != nil {
		return err
	}
	return nil
}

func (e *Equipment) ReadAll() ([]Equipment, error) {
	var equipment []Equipment
	if err := db.Find(&equipment).Error; err != nil {
		return equipment, err
	}
	return equipment, nil
}

func (e *Equipment) ReadByName() error {
	if err := db.Where("name = ?", e.Name).First(e).Error; err != nil {
		return err
	}
	return nil
}
