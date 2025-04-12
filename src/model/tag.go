package model

func (t *Tag) Create() (err error) {
	if err := db.Create(t).Error; err != nil {
		return err
	}
	return nil
}

func (t *Tag) ReadByID() error {
	if err := db.Limit(1).Find(t, t.ID).Error; err != nil {
		return err
	}
	return nil
}

func (t *Tag) ReadByName() error {
	if err := db.Where("name = ?", t.Name).Limit(1).Find(t).Error; err != nil {
		return err
	}
	return nil
}

func (t *Tag) ReadAll() ([]Tag, error) {
	var tags []Tag
	if err := db.Find(&tags).Error; err != nil {
		return tags, err
	}
	return tags, nil
}

type TagGroup struct {
	Tag   Tag
	Users []User
}

// tagごとにユーザーをグループ化
func GroupByTag(users []User) (map[string]TagGroup, error) {
	tagGroups := make(map[string]TagGroup)
	var t Tag

	// すべてのタグを取得して初期化
	tags, err := t.ReadAll()
	if err != nil {
		return nil, err
	}
	for _, tag := range tags {
		tagGroups[tag.Name] = TagGroup{
			Tag:   tag,
			Users: []User{},
		}
	}

	// 対象ユーザのIDを収集
	var userIDs []uint
	for _, u := range users {
		userIDs = append(userIDs, u.ID)
	}

	// 対象ユーザをまとめてPreloadして取得
	var enrichedUsers []User
	if err := db.Preload("Corresponds").Find(&enrichedUsers, userIDs).Error; err != nil {
		return nil, err
	}
	// log.Default().Println("enrichedUsers", enrichedUsers)

	// for _, user := range users {
	// 	var userWithTags User
	// 	if err := db.Preload("Corresponds.Tag").First(&userWithTags, user.ID).Error; err != nil {
	// 		continue
	// 	}
	// 	for _, corr := range userWithTags.Corresponds {
	// 		tagID := corr.TagID
	// 		var tag Tag
	// 		if err := db.First(&tag, tagID).Error; err != nil {
	// 			continue
	// 		}
	// 		group := tagGroups[tag.Name]
	// 		group.Users = append(group.Users, user)
	// 		tagGroups[tag.Name] = group
	// 	}
	// }

	// 対応関係からユーザをタグごとに分類
	for _, user := range enrichedUsers {
		for _, corr := range user.Corresponds {
			var tag Tag
			if err := db.First(&tag, corr.TagID).Error; err != nil {
				continue
			}
			group := tagGroups[tag.Name]
			group.Users = append(group.Users, user)
			tagGroups[tag.Name] = group
		}
	}

	return tagGroups, nil
}
