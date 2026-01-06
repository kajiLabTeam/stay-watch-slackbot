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

// ReadAllWithUsers は全イベントを Corresponds と User を含めて取得する
// NotifyByEvent などで N+1 クエリ問題を回避するために使用
func (e *Event) ReadAllWithUsers() ([]Event, error) {
	var events []Event
	if err := db.Preload("Type").Preload("Tools").Preload("Corresponds.User").Find(&events).Error; err != nil {
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

// EventGroup represents an Event with its associated Users
type EventGroup struct {
	Event Event
	Users []User
}

// GroupByEvent groups users by their associated events
func GroupByEvent(users []User) ([]EventGroup, error) {
	if len(users) == 0 {
		return []EventGroup{}, nil
	}

	// ステップ1: 全ユーザーIDを収集
	var userIDs []uint
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	// ステップ2: バッチで全 corresponds を取得（N+1 クエリ問題を回避）
	var c Correspond
	corresponds, err := c.ReadByUserIDs(userIDs)
	if err != nil {
		return nil, err
	}

	// ステップ3: メモリ上でグループ化
	eventMap := make(map[uint]*EventGroup)
	userMap := make(map[uint]User)

	// ユーザーマップを作成
	for _, user := range users {
		userMap[user.ID] = user
	}

	// イベントごとにグループ化
	for _, correspond := range corresponds {
		if _, exists := eventMap[correspond.EventID]; !exists {
			eventMap[correspond.EventID] = &EventGroup{
				Event: correspond.Event,
				Users: []User{},
			}
		}
		if user, found := userMap[correspond.UserID]; found {
			eventMap[correspond.EventID].Users = append(eventMap[correspond.EventID].Users, user)
		}
	}

	// マップをスライスに変換
	var eventGroups []EventGroup
	for _, group := range eventMap {
		eventGroups = append(eventGroups, *group)
	}

	return eventGroups, nil
}
