package model

func (tag Tag) Create() error {
	if err:= db.Create(&tag).Error; err != nil {
		return err
	}
	return nil
}

func (tag *Tag) Read() error{
	if err := db.Where(&tag).Find(&tag).Error; err != nil {
		return err
	}
	return nil
}

func ReadAllTags() (tags []Tag, err error) {
	if err := db.Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}