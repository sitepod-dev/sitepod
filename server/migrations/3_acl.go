package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db)

		authRule := "@request.auth.id != ''"
		projectRule := authRule + " && owner_id = @request.auth.id"
		projectRelRule := authRule + " && project_id.owner_id = @request.auth.id"
		previewRule := authRule + " && image_id.project_id.owner_id = @request.auth.id"

		updateRules := func(name string, rule string) error {
			collection, err := dao.FindCollectionByNameOrId(name)
			if err != nil {
				return err
			}
			collection.ListRule = types.Pointer(rule)
			collection.ViewRule = types.Pointer(rule)
			collection.CreateRule = types.Pointer(rule)
			collection.UpdateRule = types.Pointer(rule)
			collection.DeleteRule = types.Pointer(rule)
			return dao.SaveCollection(collection)
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
	}, func(db dbx.Builder) error {
		dao := daos.New(db)

		authRule := "@request.auth.id != ''"

		resetRules := func(name string) error {
			collection, err := dao.FindCollectionByNameOrId(name)
			if err != nil {
				return err
			}
			collection.ListRule = types.Pointer(authRule)
			collection.ViewRule = types.Pointer(authRule)
			collection.CreateRule = types.Pointer(authRule)
			collection.UpdateRule = types.Pointer(authRule)
			collection.DeleteRule = types.Pointer(authRule)
			return dao.SaveCollection(collection)
		}

		for _, name := range []string{"projects", "images", "plans", "deploy_events", "domains", "previews"} {
			if err := resetRules(name); err != nil {
				return err
			}
		}

		return nil
	})
}
