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
	var eventGroups []EventGroup
	eventMap := make(map[uint]*EventGroup)

	for _, user := range users {
		// Get all corresponds for this user
		correspond := Correspond{UserID: user.ID}
		corresponds, err := correspond.ReadByUserID()
		if err != nil {
			return nil, err
		}

		// Add user to each event group
		for _, c := range corresponds {
			if _, exists := eventMap[c.EventID]; !exists {
				eventMap[c.EventID] = &EventGroup{
					Event: c.Event,
					Users: []User{},
				}
			}
			eventMap[c.EventID].Users = append(eventMap[c.EventID].Users, user)
		}
	}

	// Convert map to slice
	for _, group := range eventMap {
		eventGroups = append(eventGroups, *group)
	}

	return eventGroups, nil
}
