package pollrepo

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

const VoteLimit = 100

type Repository struct {
	pool *pgxpool.Pool
}

// ** MetricsRepository wraps the base repository to add metrics **
type MetricsRepository struct {
	repo IPollRepository
}

func New(pool *pgxpool.Pool) IPollRepository {
	base := Repository{
		pool: pool,
	}

	return &MetricsRepository{
		repo: base,
	}
}
