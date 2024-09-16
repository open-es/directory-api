package database

import (
	"context"
	"log"

	db "directory/pkg/database"
	"directory/pkg/types"
)

// TODO Rename to features
type FeatureStore struct {
	store db.Pool
}

func NewFeatureStore(s db.Pool) *FeatureStore {
	return &FeatureStore{store: s}
}

func (s *FeatureStore) Create(ctx context.Context, feature types.Feature) (createdId int, err error) {
	err = s.store.QueryRow(ctx, createQuery, feature.Name, feature.Type, feature.ParentId).Scan(&createdId)
	return
}

func (s *FeatureStore) FindByID(ctx context.Context, id int) (feature *types.Feature, err error) {
	feature = &types.Feature{}
	err = s.store.QueryRow(ctx, findByIDQuery, id).Scan(&feature.Id, &feature.Name, &feature.Type, &feature.ParentId)
	return
}

func (s *FeatureStore) Update(ctx context.Context, feature types.Feature) (updatedId int, err error) {
	err = s.store.QueryRow(ctx, updateQuery, feature.Id, feature.Name, feature.Type, feature.ParentId).Scan(&updatedId)
	return
}

func (s *FeatureStore) Delete(ctx context.Context, id int) (deletedId int, err error) {
	err = s.store.QueryRow(ctx, deleteQuery, id).Scan(&deletedId)
	return
}

func (s *FeatureStore) FindRelationsByID(ctx context.Context, id int) (features []*types.Feature, featureIds []*int, err error) {
	rows, err := s.store.Query(ctx, recursiveFindByIDQuery, id)
	if err != nil {
		log.Fatalf("failed to query rows: %v", err)
	}
	for rows.Next() {
		var feature types.Feature
		err := rows.Scan(&feature.Id, &feature.Name, &feature.Type, &feature.ParentId)
		if err != nil {
			log.Fatalf("failed to scan row: %v", err)
		}
		features = append(features, &feature)
		featureIds = append(featureIds, &feature.Id)
	}

	return
}

const createQuery = "INSERT INTO directory.features (name, type, parent_id) VALUES ($1, $2, $3) RETURNING id"

const findByIDQuery = "SELECT d.id, d.name, d.type, d.parent_id FROM directory.features d WHERE id = $1"

const updateQuery = "UPDATE directory.features SET name = $2, type = $3, parent_id = $4 WHERE id = $1 RETURNING id"

const deleteQuery = "DELETE FROM directory.features WHERE id = $1 RETURNING id"

// WARNING Can this ever be circular?
// If a parent references a child incorrectly -> child references parent correctly
const recursiveFindByIDQuery = `
WITH RECURSIVE ParentCTE AS (
		-- Start with the given record and find its children
		SELECT internal_id, id, name, type, parent_id
		FROM directory.features
		WHERE internal_id = 1
		
		UNION ALL
		
		-- Find all parents of the current record
		SELECT loc.internal_id, loc.id, loc.name, loc.type, loc.parent_id
		FROM directory.features loc
		JOIN ParentCTE p ON loc.internal_id = p.parent_id
	),
	ChildCTE AS (
		-- Start with the given record and find its children
		SELECT internal_id, id, name, type, parent_id
		FROM directory.features
		WHERE internal_id = 1
		
		UNION ALL
		
		-- Find all children of the current record
		SELECT loc.internal_id, loc.id, loc.name, loc.type, loc.parent_id
		FROM directory.features loc
		JOIN ChildCTE c ON loc.parent_id = c.internal_id
	)
	-- Combine results from both ParentCTE and ChildCTE
	SELECT * FROM ParentCTE
	UNION
	SELECT * FROM ChildCTE
	ORDER BY id;
`
