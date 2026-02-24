package database

import (
	"database/sql"
	"fmt"
	"log"
)

func RunMigrations(db *sql.DB) error {
	migrations := []string{
		createUsersTable,
		createSeriesTable,
		createEpisodesTable,
		createUnlocksTable,
		createTransactionsTable,
		createViewsTable,
		createBansTable,
		createUserRolesTable,
		createRefreshTokensTable,
		createAdValidationsTable,
		createIndexes,
		createAdditionalIndexes,
		createFavoritesTable,
		createWatchProgressSupport,
		createDeviceTokensTable,
		// Multi-tenant
		createProducersTable,
		alterSeriesAddProducerID,
		alterUserRolesAddProducerRoles,
		// Onboarding + invitations
		alterProducersAddStatus,
		createInvitationsTable,
		// Equipo multi-usuario por tenant
		createProducerMembersTable,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	log.Println("✅ Database migrations completed successfully")
	return nil
}

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    firebase_uid VARCHAR(255) UNIQUE NOT NULL,
    coin_balance INTEGER DEFAULT 0 CHECK (coin_balance >= 0),
    is_premium BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createSeriesTable = `
CREATE TABLE IF NOT EXISTS series (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    horizontal_poster VARCHAR(500),
    vertical_poster VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createEpisodesTable = `
CREATE TABLE IF NOT EXISTS episodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    series_id UUID NOT NULL REFERENCES series(id) ON DELETE CASCADE,
    episode_number INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    video_id_bunny VARCHAR(255),
    duration INTEGER DEFAULT 0,
    is_free BOOLEAN DEFAULT FALSE,
    price_coins INTEGER DEFAULT 0 CHECK (price_coins >= 0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(series_id, episode_number)
);
`

const createUnlocksTable = `
CREATE TABLE IF NOT EXISTS unlocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    episode_id UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    method VARCHAR(20) NOT NULL CHECK (method IN ('COIN', 'AD', 'SUB')),
    unlocked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, episode_id)
);
`

const createTransactionsTable = `
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('unlock', 'purchase', 'gift', 'ad_reward')),
    amount INTEGER NOT NULL,
    episode_id UUID REFERENCES episodes(id) ON DELETE SET NULL,
    method VARCHAR(20) NOT NULL CHECK (method IN ('COIN', 'AD', 'SUB', 'GIFT')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createViewsTable = `
CREATE TABLE IF NOT EXISTS views (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    episode_id UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    watched_seconds INTEGER DEFAULT 0,
    completed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createBansTable = `
CREATE TABLE IF NOT EXISTS bans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason TEXT,
    banned_by UUID REFERENCES users(id) ON DELETE SET NULL,
    expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createUserRolesTable = `
CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'moderator', 'user')),
    granted_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role)
);
`

const createRefreshTokensTable = `
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createAdValidationsTable = `
CREATE TABLE IF NOT EXISTS ad_validations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ad_id VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    episode_id UUID REFERENCES episodes(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createIndexes = `
CREATE INDEX IF NOT EXISTS idx_users_firebase_uid ON users(firebase_uid);
CREATE INDEX IF NOT EXISTS idx_episodes_series_id ON episodes(series_id);
CREATE INDEX IF NOT EXISTS idx_episodes_is_free ON episodes(is_free);
CREATE INDEX IF NOT EXISTS idx_unlocks_user_id ON unlocks(user_id);
CREATE INDEX IF NOT EXISTS idx_unlocks_episode_id ON unlocks(episode_id);
CREATE INDEX IF NOT EXISTS idx_series_is_active ON series(is_active);
`

const createAdditionalIndexes = `
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_views_episode_id ON views(episode_id);
CREATE INDEX IF NOT EXISTS idx_views_user_id ON views(user_id);
CREATE INDEX IF NOT EXISTS idx_views_created_at ON views(created_at);
CREATE INDEX IF NOT EXISTS idx_bans_user_id ON bans(user_id);
CREATE INDEX IF NOT EXISTS idx_bans_is_active ON bans(is_active);
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role ON user_roles(role);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_revoked ON refresh_tokens(revoked);
CREATE INDEX IF NOT EXISTS idx_ad_validations_ad_id ON ad_validations(ad_id);
CREATE INDEX IF NOT EXISTS idx_ad_validations_user_id ON ad_validations(user_id);
CREATE INDEX IF NOT EXISTS idx_ad_validations_created_at ON ad_validations(created_at);
`

const createFavoritesTable = `
CREATE TABLE IF NOT EXISTS favorites (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    series_id  UUID NOT NULL REFERENCES series(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, series_id)
);
CREATE INDEX IF NOT EXISTS idx_favorites_user_id ON favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_favorites_series_id ON favorites(series_id);
`

// createDeviceTokensTable crea la tabla de tokens FCM para notificaciones push.
const createDeviceTokensTable = `
CREATE TABLE IF NOT EXISTS device_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(512) UNIQUE NOT NULL,
    platform   VARCHAR(16) NOT NULL CHECK (platform IN ('android','ios')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_device_tokens_user_id ON device_tokens(user_id);
`

// createWatchProgressSupport añade la columna updated_at a views y crea el índice único parcial
// necesario para el UPSERT de progreso de visualización (UpdateWatchProgress).
const createWatchProgressSupport = `
ALTER TABLE views ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
CREATE UNIQUE INDEX IF NOT EXISTS idx_views_user_episode
    ON views (user_id, episode_id)
    WHERE user_id IS NOT NULL;
`

// ─── Multi-tenant ──────────────────────────────────────────────────────────────

// createProducersTable crea la tabla de productores de contenido.
// Cada productor está vinculado a un usuario con rol 'producer' o 'super_admin'.
const createProducersTable = `
CREATE TABLE IF NOT EXISTS producers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    slug        VARCHAR(100) UNIQUE NOT NULL,
    logo_url    VARCHAR(500),
    description TEXT,
    is_active   BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_producers_user_id ON producers(user_id);
CREATE INDEX IF NOT EXISTS idx_producers_slug ON producers(slug);
`

// alterSeriesAddProducerID añade la FK producer_id a la tabla series.
// Nullable: las series sin producer_id son "contenido de plataforma" visible para super_admin.
const alterSeriesAddProducerID = `
ALTER TABLE series ADD COLUMN IF NOT EXISTS producer_id UUID REFERENCES producers(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_series_producer_id ON series(producer_id);
`

// alterUserRolesAddProducerRoles amplía el CHECK constraint de user_roles para soportar
// los nuevos roles 'producer' y 'super_admin'.
// Primero elimina el constraint viejo (si existe) y luego lo recrea ampliado.
const alterUserRolesAddProducerRoles = `
ALTER TABLE user_roles DROP CONSTRAINT IF EXISTS user_roles_role_check;
ALTER TABLE user_roles ADD CONSTRAINT user_roles_role_check
    CHECK (role IN ('admin', 'moderator', 'user', 'producer', 'super_admin'));
`

// alterProducersAddStatus agrega el campo status para el flujo de aprobación de tenants.
// pending = esperando aprobación del super_admin
// active   = productor activo con acceso completo
// suspended = acceso suspendido temporalmente
const alterProducersAddStatus = `
ALTER TABLE producers ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'pending'
    CHECK (status IN ('pending', 'active', 'suspended'));
`

// createInvitationsTable almacena links de invitación para añadir colaboradores a un tenant.
// El tenant admin genera un token único con rol y expiración, lo comparte vía link.
const createInvitationsTable = `
CREATE TABLE IF NOT EXISTS invitations (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token        VARCHAR(64) UNIQUE NOT NULL,
    producer_id  UUID NOT NULL REFERENCES producers(id) ON DELETE CASCADE,
    role         VARCHAR(50) NOT NULL DEFAULT 'producer',
    created_by   UUID REFERENCES users(id),
    expires_at   TIMESTAMP NOT NULL,
    used_at      TIMESTAMP,
    used_by      UUID REFERENCES users(id),
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_invitations_token       ON invitations(token);
CREATE INDEX IF NOT EXISTS idx_invitations_producer_id ON invitations(producer_id);
`

// createProducerMembersTable permite que múltiples usuarios pertenezcan al mismo tenant.
// El dueño sigue siendo el user_id en la tabla producers;
// los colaboradores invitados se almacenan aquí.
const createProducerMembersTable = `
CREATE TABLE IF NOT EXISTS producer_members (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    producer_id UUID NOT NULL REFERENCES producers(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        VARCHAR(50) NOT NULL DEFAULT 'producer',
    joined_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (producer_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_producer_members_user_id     ON producer_members(user_id);
CREATE INDEX IF NOT EXISTS idx_producer_members_producer_id ON producer_members(producer_id);
`

