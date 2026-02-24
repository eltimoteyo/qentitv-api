package app

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

// resolveProducerSlug convierte un slug de productor a su UUID.
// Devuelve nil (sin error) si slug es vacío → comportamiento "sin filtro".
// Devuelve nil + error si el slug no existe en la BD.
func resolveProducerSlug(ctx context.Context, db *sql.DB, slug string) (*uuid.UUID, error) {
	if slug == "" {
		return nil, nil
	}
	var id uuid.UUID
	err := db.QueryRowContext(ctx,
		`SELECT id FROM producers WHERE slug = $1 AND status = 'active' LIMIT 1`, slug,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return nil, nil // slug desconocido → ignorar filtro (mejor que 404)
	}
	if err != nil {
		return nil, err
	}
	return &id, nil
}
