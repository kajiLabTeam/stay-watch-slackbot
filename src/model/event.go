package model

func (e *Event) Create() error {
	if err := db.Create(e).Error; err != nil {
		return err
	}
	return nil
}

func (e *Event) ReadByID() error {
	if err := db.Preload("Type").Preload("Tools").First(e, e.ID).Error; err != nil {
		return err
	}
	return nil
}

func (e *Event) ReadByName() error {
	if err := db.Preload("Type").Preload("Tools").Where("name = ?", e.Name).First(e).Error; err != nil {
		return err
	}
	return nil
}

func (e *Event) ReadAll() ([]Event, error) {
	var events []Event
	if err := db.Preload("Type").Preload("Tools").Find(&events).Error; err != nil {
		return events, err
	}
	return events, nil
}

func (e *Event) Update() error {
	if err := db.Save(e).Error; err != nil {
		return err
	}
	return nil
}

func (e *Event) Delete() error {
	if err := db.Delete(e).Error; err != nil {
		return err
	}
	return nil
}
