package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		authRule := "@request.auth.id != ''"
		projectRule := authRule + " && owner_id = @request.auth.id"
		projectRelRule := authRule + " && project_id.owner_id = @request.auth.id"
		previewRule := authRule + " && image_id.project_id.owner_id = @request.auth.id"

		updateRules := func(name string, rule string) error {
			collection, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				return err
			}
			collection.ListRule = ptrStr(rule)
			collection.ViewRule = ptrStr(rule)
			collection.CreateRule = ptrStr(rule)
			collection.UpdateRule = ptrStr(rule)
			collection.DeleteRule = ptrStr(rule)
			return app.Save(collection)
		}

		if err := updateRules("projects", projectRule); err != nil {
			return err
		}
		if err := updateRules("images", projectRelRule); err != nil {
			return err
		}
		if err := updateRules("plans", projectRelRule); err != nil {
			return err
		}
		if err := updateRules("deploy_events", projectRelRule); err != nil {
			return err
		}
		if err := updateRules("domains", projectRelRule); err != nil {
			return err
		}
		if err := updateRules("previews", previewRule); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		authRule := "@request.auth.id != ''"

		resetRules := func(name string) error {
			collection, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				return err
			}
			collection.ListRule = ptrStr(authRule)
			collection.ViewRule = ptrStr(authRule)
			collection.CreateRule = ptrStr(authRule)
			collection.UpdateRule = ptrStr(authRule)
			collection.DeleteRule = ptrStr(authRule)
			return app.Save(collection)
		}

		for _, name := range []string{"projects", "images", "plans", "deploy_events", "domains", "previews"} {
			if err := resetRules(name); err != nil {
				return err
			}
		}

		return nil
	})
}
