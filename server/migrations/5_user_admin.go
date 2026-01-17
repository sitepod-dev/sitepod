package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models/schema"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db)

		users, err := dao.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		users.Schema.AddField(&schema.SchemaField{
			Name:     "is_admin",
			Type:     schema.FieldTypeBool,
			Required: false,
		})

		return dao.SaveCollection(users)
	}, func(db dbx.Builder) error {
		dao := daos.New(db)

		users, err := dao.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		users.Schema.RemoveField("is_admin")
		return dao.SaveCollection(users)
	})
}
