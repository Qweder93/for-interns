package clients_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cleanmasters"
	"cleanmasters/clients"
	"cleanmasters/database/dbtesting"
)

func TestAccounts(t *testing.T) {
	dbtesting.Run(t, func(ctx context.Context, t *testing.T, db cleanmasters.DB) {
		repo := db.Clients()

		id, err := repo.Register(ctx, "228")
		require.NoError(t, err)

		clientCheck, err := repo.Get(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, clientCheck.ID, id)

		id1 := uuid.New()

		client := clients.Client{
			ID:        id1,
			Phone:     "0930000000",
			FirstName: "Aslan",
			LastName:  "Maslan",
			Email:     "am@qwe.com",
		}

		err = repo.Add(ctx, client)
		require.NoError(t, err)

		clientCheck, err = repo.Get(ctx, id1)
		require.NoError(t, err)
		assert.Equal(t, clientCheck.ID, client.ID)
		assert.Equal(t, clientCheck.Phone, client.Phone)
		assert.Equal(t, clientCheck.FirstName, client.FirstName)
		assert.Equal(t, clientCheck.LastName, client.LastName)
		assert.Equal(t, clientCheck.Email, client.Email)

		id2 := uuid.New()
		client2 := clients.Client{
			ID:        id2,
			Phone:     "14882288",
			FirstName: "Baslan",
			LastName:  "Haslan",
			Email:     "bh@qwe.com",
		}

		err = repo.Add(ctx, client2)
		require.NoError(t, err)

		clientCheck2, err := repo.Get(ctx, id2)
		require.NoError(t, err)
		assert.Equal(t, clientCheck2.ID, client2.ID)
		assert.Equal(t, clientCheck2.Phone, client2.Phone)
		assert.Equal(t, clientCheck2.FirstName, client2.FirstName)
		assert.Equal(t, clientCheck2.LastName, client2.LastName)
		assert.Equal(t, clientCheck2.Email, client2.Email)

		err = repo.Update(ctx, clients.Client{
			ID:        id1,
			Email:     "new_poshta_am@qwe.com",
			Phone:     "5051228",
			FirstName: "Aslan",
			LastName:  "Maslanovich",
		})
		require.NoError(t, err)

		clientCheck, err = repo.GetByPhone(ctx, "5051228")
		require.NoError(t, err)
		assert.Equal(t, clientCheck.ID, id1)

		list, err := repo.List(ctx)
		require.NoError(t, err)
		assert.NotNil(t, list)
		assert.Equal(t, id1, list[2].ID)

		err = repo.Delete(ctx, id1)
		require.NoError(t, err)

		_, err = repo.Get(ctx, id1)
		assert.Error(t, err)
	})
}
