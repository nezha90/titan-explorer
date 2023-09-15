package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

var tableNameFilStorage = "fil_storage"

func AddFilStorages(ctx context.Context, storages []*model.FilStorage) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s ( provider, sector_num, cost, message_cid, piece_cid, payload_cid, deal_id, path, piece_size, start_height, end_height, created_at, updated_at)
			VALUES ( :provider, :sector_num, :cost, :message_cid, :piece_cid, :payload_cid, :deal_id, :path, :piece_size, :start_height, :end_height, :created_at, :updated_at);`, tableNameFilStorage,
	), storages)
	return err
}

func ListFilStorages(ctx context.Context, path string, option QueryOption) ([]*model.FilStorage, int64, error) {
	var args []interface{}
	var total int64
	var out []*model.FilStorage

	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 50
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}

	args = append(args, path)

	err := DB.GetContext(ctx, &total, fmt.Sprintf(
		`SELECT count(*) FROM %s WHERE path = ?`, tableNameFilStorage,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s WHERE path = ? LIMIT %d OFFSET %d`, tableNameFilStorage, limit, offset,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}
