package model

func (t *Tool) Create() error {
	if err := db.Create(t).Error; err != nil {
		return err
	}
	return nil
}

func (t *Tool) ReadByID() error {
	if err := db.First(t, t.ID).Error; err != nil {
		return err
	}
	return nil
}

func (t *Tool) ReadByName() error {
	if err := db.Where("name = ?", t.Name).First(t).Error; err != nil {
		return err
	}
	return nil
}

func (t *Tool) ReadAll() ([]Tool, error) {
	var tools []Tool
	if err := db.Find(&tools).Error; err != nil {
		return tools, err
	}
	return tools, nil
}
