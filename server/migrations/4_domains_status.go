package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		domains, err := app.FindCollectionByNameOrId("domains")
		if err != nil {
			return err
		}

		// Update status field values - in v0.36, we need to modify the field
		field := domains.Fields.GetByName("status")
		if field != nil {
			if selectField, ok := field.(*core.SelectField); ok {
				selectField.Values = []string{"pending", "active"}
			}
		}

		if err := app.Save(domains); err != nil {
			return err
		}

		// Update existing records
		records, err := app.FindRecordsByFilter(
			"domains",
			"status = 'verified'",
			"",
			1000,
			0,
			nil,
		)
		if err != nil {
			return nil // No records found is fine
		}

		for _, record := range records {
			record.Set("status", "active")
			if err := app.Save(record); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		domains, err := app.FindCollectionByNameOrId("domains")
		if err != nil {
			return err
		}

		field := domains.Fields.GetByName("status")
		if field != nil {
			if selectField, ok := field.(*core.SelectField); ok {
				selectField.Values = []string{"pending", "verified", "active"}
			}
		}

		return app.Save(domains)
	})
}
