package model

func (user User) Create() (err error) {
	if err := db.Create(&user).Error; err != nil {
		return err
	}
	return nil
}

func (user *User) Read() (err error) {
	if err := db.Where(&user).Find(&user).Error; err != nil {
		return err
	}
	return nil
}