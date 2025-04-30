package pollrepo

import (
	"context"
	"errors"
	"fmt"

	"pollapp/internal/repository/model"
	"pollapp/internal/service/domain"

	"github.com/jackc/pgx/v5"
)

// Create creates a new poll with options and tags
func (r Repository) Create(ctx context.Context, req model.Poll) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var pollID int
	err = tx.QueryRow(ctx, `
		INSERT INTO polls (title, created_at, updated_at) 
		VALUES ($1, NOW(), NOW()) 
		RETURNING id
	`, req.Title).Scan(&pollID)
	if err != nil {
		return fmt.Errorf("insert poll: %w", err)
	}

	for _, option := range req.Options {
		_, err = tx.Exec(ctx, `
			INSERT INTO poll_options (poll_id, option_text, created_at, updated_at) 
			VALUES ($1, $2, NOW(), NOW())
		`, pollID, option)
		if err != nil {
			return fmt.Errorf("insert option: %w", err)
		}
	}

	for _, tag := range req.Tags {
		var tagID int
		err = tx.QueryRow(ctx, `
			INSERT INTO tags (name, created_at) 
			VALUES ($1, NOW()) 
			ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, tag).Scan(&tagID)
		if err != nil {
			return fmt.Errorf("insert tag: %w", err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO poll_tags (poll_id, tag_id) 
			VALUES ($1, $2)
		`, pollID, tagID)
		if err != nil {
			return fmt.Errorf("insert poll_tag: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// Get retrieves a list of polls based on filter criteria
func (r Repository) Get(ctx context.Context, filter domain.GetPollsParams) ([]domain.Poll, error) {
	query := `
		WITH excluded_polls AS (
			SELECT poll_id FROM votes WHERE user_id = $1
			UNION
			SELECT poll_id FROM skips WHERE user_id = $1
		)
		SELECT p.id, p.title, p.created_at
		FROM polls p
		WHERE p.id NOT IN (SELECT poll_id FROM excluded_polls)
	`

	args := []interface{}{filter.UserID}
	paramCount := 1
	if filter.Tag != "" {
		paramCount++
		query += `
			AND p.id IN (
				SELECT pt.poll_id 
				FROM poll_tags pt 
				JOIN tags t ON pt.tag_id = t.id 
				WHERE t.name = $` + fmt.Sprintf("%d", paramCount) + `
			)
		`
		args = append(args, filter.Tag)
	}

	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}

	offset := (filter.Page - 1) * filter.Limit

	paramCount++
	limitParam := paramCount
	paramCount++
	offsetParam := paramCount

	query += fmt.Sprintf(`
		ORDER BY p.created_at DESC
		LIMIT $%d OFFSET $%d
	`, limitParam, offsetParam)

	args = append(args, filter.Limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query polls: %w", err)
	}
	defer rows.Close()

	polls := []domain.Poll{}
	for rows.Next() {
		var poll domain.Poll
		err = rows.Scan(&poll.ID, &poll.Title, &poll.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan poll: %w", err)
		}

		options, err := r.getPollOptions(ctx, int(poll.ID))
		if err != nil {
			return nil, fmt.Errorf("get poll options: %w", err)
		}
		poll.Options = options

		tags, err := r.getPollTags(ctx, int(poll.ID))
		if err != nil {
			return nil, fmt.Errorf("get poll tags: %w", err)
		}
		poll.Tags = tags

		polls = append(polls, poll)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return polls, nil
}

// GetByID retrieves a poll by ID
func (r Repository) GetByID(ctx context.Context, id uint) (domain.Poll, error) {
	var poll model.Poll

	err := r.pool.QueryRow(ctx, `
		SELECT id, title, created_at
		FROM polls
		WHERE id = $1
	`, id).Scan(&poll.ID, &poll.Title, &poll.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Poll{}, errors.New("poll not found")
		}
		return domain.Poll{}, fmt.Errorf("query poll: %w", err)
	}

	options, err := r.getPollOptions(ctx, int(poll.ID))
	if err != nil {
		return domain.Poll{}, fmt.Errorf("get poll options: %w", err)
	}
	poll.Options = options

	tags, err := r.getPollTags(ctx, int(poll.ID))
	if err != nil {
		return domain.Poll{}, fmt.Errorf("get poll tags: %w", err)
	}
	poll.Tags = tags

	return poll.ToDomain(), nil
}

// Vote records a user's vote on a poll
func (r Repository) Vote(ctx context.Context, pollID uint, vote model.Vote) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	voteCount, err := r.checkDailyVoteLimitTx(ctx, tx, uint(vote.UserID))
	if err != nil {
		return fmt.Errorf("check daily vote limit: %w", err)
	}

	if voteCount >= VoteLimit {
		return errors.New("daily vote limit exceeded")
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO votes (poll_id, user_id, option_index, created_at) 
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (poll_id, user_id) DO NOTHING
	`, pollID, vote.UserID, vote.OptionIndex)
	if err != nil {
		return fmt.Errorf("insert vote: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO user_daily_votes (user_id, vote_date, vote_count) 
		VALUES ($1, CURRENT_DATE, 1)
		ON CONFLICT (user_id, vote_date) 
		DO UPDATE SET vote_count = user_daily_votes.vote_count + 1
	`, vote.UserID)
	if err != nil {
		return fmt.Errorf("update daily vote count: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// Skip records a user's decision to skip a poll
func (r Repository) Skip(ctx context.Context, pollID uint, skip domain.SkipRequest) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO skips (poll_id, user_id, created_at) 
		VALUES ($1, $2, NOW())
		ON CONFLICT (poll_id, user_id) DO NOTHING
	`, pollID, skip.UserID)
	if err != nil {
		return fmt.Errorf("insert skip: %w", err)
	}

	return nil
}

// GetStats retrieves statistics for a poll
func (r Repository) GetStats(ctx context.Context, pollID uint) (*domain.PollStats, error) {
	options, err := r.getPollOptions(ctx, int(pollID))
	if err != nil {
		return nil, fmt.Errorf("get poll options: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT option_index, COUNT(*)
		FROM votes
		WHERE poll_id = $1
		GROUP BY option_index
	`, pollID)
	if err != nil {
		return nil, fmt.Errorf("query poll stats: %w", err)
	}
	defer rows.Close()

	optionCounts := make(map[int]int)
	for rows.Next() {
		var idx, count int
		err = rows.Scan(&idx, &count)
		if err != nil {
			return nil, fmt.Errorf("scan option votes: %w", err)
		}
		optionCounts[idx] = count
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	stats := &domain.PollStats{
		PollID: int(pollID),
		Votes:  make([]domain.OptionStat, len(options)),
	}

	for i, opt := range options {
		stats.Votes[i] = domain.OptionStat{
			Option: opt,
			Count:  optionCounts[i],
		}
	}

	return stats, nil
}

// CheckDailyVoteLimit checks if a user has reached their daily vote limit
func (r Repository) CheckDailyVoteLimit(ctx context.Context, userID uint) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT vote_count FROM user_daily_votes
		WHERE user_id = $1 AND vote_date = CURRENT_DATE
	`, userID).Scan(&count)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("query vote count: %w", err)
	}

	return count, nil
}

// checkDailyVoteLimitTx checks if a user has reached their daily vote limit within a transaction
func (r Repository) checkDailyVoteLimitTx(ctx context.Context, tx pgx.Tx, userID uint) (int, error) {
	var count int
	err := tx.QueryRow(ctx, `
		SELECT vote_count FROM user_daily_votes
		WHERE user_id = $1 AND vote_date = CURRENT_DATE
	`, userID).Scan(&count)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("query vote count: %w", err)
	}

	return count, nil
}

// getPollOptions retrieves a list of options for a poll
func (r Repository) getPollOptions(ctx context.Context, pollID int) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT option_text
		FROM poll_options
		WHERE poll_id = $1
		ORDER BY id
	`, pollID)
	if err != nil {
		return nil, fmt.Errorf("query poll options: %w", err)
	}
	defer rows.Close()

	options := []string{}
	for rows.Next() {
		var option string
		err = rows.Scan(&option)
		if err != nil {
			return nil, fmt.Errorf("scan option: %w", err)
		}
		options = append(options, option)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return options, nil
}

// getPollTags retrieves tags for a poll
func (r Repository) getPollTags(ctx context.Context, pollID int) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT t.name
		FROM tags t
		JOIN poll_tags pt ON t.id = pt.tag_id
		WHERE pt.poll_id = $1
	`, pollID)
	if err != nil {
		return nil, fmt.Errorf("query poll tags: %w", err)
	}
	defer rows.Close()

	tags := []string{}
	for rows.Next() {
		var tag string
		err = rows.Scan(&tag)
		if err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return tags, nil
}

// GetAllWithTags retrieves all polls with their tags from the database
func (r Repository) GetAllWithTags(ctx context.Context) ([]domain.Poll, error) {
	query := `
		SELECT p.id, p.title, p.created_at
		FROM polls p
		ORDER BY p.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query polls: %w", err)
	}
	defer rows.Close()

	polls := []domain.Poll{}
	for rows.Next() {
		var poll domain.Poll
		err = rows.Scan(&poll.ID, &poll.Title, &poll.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan poll: %w", err)
		}

		options, err := r.getPollOptions(ctx, int(poll.ID))
		if err != nil {
			return nil, fmt.Errorf("get poll options: %w", err)
		}
		poll.Options = options

		tags, err := r.getPollTags(ctx, int(poll.ID))
		if err != nil {
			return nil, fmt.Errorf("get poll tags: %w", err)
		}
		poll.Tags = tags

		polls = append(polls, poll)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return polls, nil
}
