package model

func (eu *EventUser) Create() error {
	if err := db.Create(eu).Error; err != nil {
		return err
	}
	return nil
}

func (eu *EventUser) ReadByID() error {
	if err := db.Preload("Event").Preload("User").First(eu, eu.ID).Error; err != nil {
		return err
	}
	return nil
}

func (eu *EventUser) ReadByEventID() ([]EventUser, error) {
	var eventUsers []EventUser
	if err := db.Preload("Event").Preload("User").Where("event_id = ?", eu.EventID).Find(&eventUsers).Error; err != nil {
		return eventUsers, err
	}
	return eventUsers, nil
}

func (eu *EventUser) ReadByUserID() ([]EventUser, error) {
	var eventUsers []EventUser
	if err := db.Preload("Event").Preload("User").Where("user_id = ?", eu.UserID).Find(&eventUsers).Error; err != nil {
		return eventUsers, err
	}
	return eventUsers, nil
}

// ReadByUserIDs は複数の UserID から EventUser をバッチで取得する
// GroupByEvent などで N+1 クエリ問題を回避するために使用
func (eu *EventUser) ReadByUserIDs(userIDs []uint) ([]EventUser, error) {
	var eventUsers []EventUser
	if err := db.Preload("Event").Where("user_id IN ?", userIDs).Find(&eventUsers).Error; err != nil {
		return nil, err
	}
	return eventUsers, nil
}
