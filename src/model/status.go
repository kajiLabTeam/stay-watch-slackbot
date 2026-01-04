package model

func (s *Status) Create() error {
	if err := db.Create(s).Error; err != nil {
		return err
	}
	return nil
}

func (s *Status) ReadByID() error {
	if err := db.First(s, s.ID).Error; err != nil {
		return err
	}
	return nil
}

func (s *Status) ReadByName() error {
	if err := db.Where("name = ?", s.Name).First(s).Error; err != nil {
		return err
	}
	return nil
}

func (s *Status) ReadAll() ([]Status, error) {
	var statuses []Status
	if err := db.Find(&statuses).Error; err != nil {
		return statuses, err
	}
	return statuses, nil
}
