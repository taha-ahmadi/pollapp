package pollrepo

import "github.com/jackc/pgx/v5/pgxpool"

const VoteLimit = 100

// Repository implements the PollRepository interface for PostgreSQL
type Repository struct {
	pool *pgxpool.Pool
}

// New creates a new Repository
func New(pool *pgxpool.Pool) IPollRepository {
	return Repository{
		pool: pool,
	}
}
