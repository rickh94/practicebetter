package db

import (
	"context"

	"github.com/nrednav/cuid2"
)

func (q *Queries) GetOrCreateUser(ctx context.Context, email string) (User, error) {
	user, err := q.GetUserByEmail(ctx, email)
	if err != nil {
		user, err = q.CreateUser(ctx, CreateUserParams{
			ID:       cuid2.Generate(),
			Email:    email,
			Fullname: "",
		})
	}
	return user, err
}
