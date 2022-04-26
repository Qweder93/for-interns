package managers_test

import (
	"cleanmasters"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cleanmasters/adminportal/managers"
	"cleanmasters/database/dbtesting"
)

func TestAccounts(t *testing.T) {
	dbtesting.Run(t, func(ctx context.Context, t *testing.T, db cleanmasters.DB) {
		repo := db.Managers()

		id := uuid.New()
		created := time.Now()
		passwordHash := []byte("qwerty123")

		manager := managers.Manager{
			ID:           id,
			FirstName:    "Aslan",
			LastName:     "Maslan",
			Email:        "am@qwe.com",
			PasswordHash: passwordHash,
			CreatedAt:    created,
		}

		err := repo.Add(ctx, manager)
		require.NoError(t, err)

		managerCheck, err := repo.Get(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, managerCheck.ID, manager.ID)
		assert.Equal(t, managerCheck.FirstName, manager.FirstName)
		assert.Equal(t, managerCheck.LastName, manager.LastName)
		assert.Equal(t, managerCheck.Email, manager.Email)
		assert.Equal(t, managerCheck.PasswordHash, manager.PasswordHash)

		id2 := uuid.New()
		manager2 := managers.Manager{
			ID:           id2,
			FirstName:    "Baslan",
			LastName:     "Haslan",
			Email:        "bh@qwe.com",
			PasswordHash: passwordHash,
			CreatedAt:    created,
		}

		err = repo.Add(ctx, manager2)
		require.NoError(t, err)

		managerCheck, err = repo.GetByEmail(ctx, "bh@qwe.com")
		require.NoError(t, err)
		assert.Equal(t, managerCheck.ID, manager2.ID)

		managerCheck2, err := repo.Get(ctx, id2)
		require.NoError(t, err)
		assert.Equal(t, managerCheck2.ID, manager2.ID)
		assert.Equal(t, managerCheck2.FirstName, manager2.FirstName)
		assert.Equal(t, managerCheck2.LastName, manager2.LastName)
		assert.Equal(t, managerCheck2.Email, manager2.Email)
		assert.Equal(t, managerCheck2.PasswordHash, manager2.PasswordHash)

		manager = managers.Manager{
			ID:           id,
			FirstName:    "QWERTT",
			LastName:     "Maslan",
			Email:        "am@qwe.com",
			PasswordHash: passwordHash,
			CreatedAt:    created,
		}

		err = repo.Update(ctx, manager)
		require.NoError(t, err)

		managerCheck, err = repo.GetByEmail(ctx, "am@qwe.com")
		require.NoError(t, err)
		assert.Equal(t, manager.FirstName, managerCheck.FirstName)
		assert.Equal(t, managerCheck.ID, manager.ID)

		list, err := repo.List(ctx)
		require.NoError(t, err)
		assert.NotNil(t, list)
		assert.Equal(t, id, list[1].ID)
		assert.Equal(t, id2, list[0].ID)

		err = repo.Remove(ctx, id)
		require.NoError(t, err)

		_, err = repo.Get(ctx, id)
		assert.Error(t, err)
	})
}
