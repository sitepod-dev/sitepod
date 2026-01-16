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

		domains, err := dao.FindCollectionByNameOrId("domains")
		if err != nil {
			return err
		}

		field := domains.Schema.GetFieldByName("status")
		if field != nil {
			if opts, ok := field.Options.(*schema.SelectOptions); ok {
				opts.Values = []string{"pending", "active"}
			}
		}

		if err := dao.SaveCollection(domains); err != nil {
			return err
		}

		records, err := dao.FindRecordsByFilter(
			"domains",
			"status = 'verified'",
			"",
			1000,
			0,
			nil,
		)
		if err != nil {
			return nil
		}

		for _, record := range records {
			record.Set("status", "active")
			dao.SaveRecord(record)
		}

		return nil
	}, func(db dbx.Builder) error {
		dao := daos.New(db)

		domains, err := dao.FindCollectionByNameOrId("domains")
		if err != nil {
			return err
		}

		field := domains.Schema.GetFieldByName("status")
		if field != nil {
			if opts, ok := field.Options.(*schema.SelectOptions); ok {
				opts.Values = []string{"pending", "verified", "active"}
			}
		}

		return dao.SaveCollection(domains)
	})
}
