package model

func (c Correspondence) Create() error {
	if err := db.Create(&c).Error; err != nil {
		return err
	}
	return nil
}

func (c *Correspondence) Read() (err error){
	if err := db.Where(&c).Find(&c).Error; err != nil {
		return err
	}
	return nil
}

func ReadAllCorrespondences() (correspondences []Correspondence, err error) {
	if err := db.Find(&correspondences).Error; err != nil {
		return nil, err
	}
	return correspondences, nil
}