package infrastructure

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zjrosen/perles/internal/beads/application"
	"github.com/zjrosen/perles/internal/testutil"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

func TestDetectSchema_ReturnsFullWhenGasTownPresent(t *testing.T) {
	db := testutil.NewTestDB(t)
	defer db.Close()

	schema, err := detectSchema(db)
	require.NoError(t, err)
	require.Equal(t, application.SchemaFull, schema)
}

func TestDetectSchema_ReturnsClassicWhenAbsent(t *testing.T) {
	db := testutil.NewClassicTestDB(t)
	defer db.Close()

	schema, err := detectSchema(db)
	require.NoError(t, err)
	require.Equal(t, application.SchemaClassic, schema)
}

func TestVersion_ClassicSchema_ReturnsNoSuchTableError(t *testing.T) {
	db := testutil.NewClassicTestDB(t)
	defer db.Close()

	var version string
	err := db.QueryRow("SELECT value FROM metadata WHERE key = 'bd_version'").Scan(&version)
	require.Error(t, err)
	require.ErrorContains(t, err, "no such table")
}

func TestDetectSchema_IssuesTableMissing(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = detectSchema(db)
	require.Error(t, err)
}
