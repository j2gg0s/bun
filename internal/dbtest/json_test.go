package dbtest_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func TestJSON(t *testing.T) {
	type User struct {
		ID    int
		Name  string
		Attrs map[string]interface{} `bun:"type:json"`
	}

	testEachDB(t, func(t *testing.T, db *bun.DB) {
		ctx := context.Background()

		_, err := db.NewDropTable().Model((*User)(nil)).IfExists().Exec(ctx)
		require.NoError(t, err)
		_, err = db.NewCreateTable().Model((*User)(nil)).Exec(ctx)
		require.NoError(t, err)

		user := User{
			Name: "j2gg0s",
			Attrs: map[string]interface{}{
				"hello": "world\nworld",
			},
		}
		_, err = db.NewInsert().Model(&user).Exec(context.Background())
		require.NoError(t, err)
	})
}
