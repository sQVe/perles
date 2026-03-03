package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTestDB_CreatesSchema(t *testing.T) {
	db := NewTestDB(t)
	defer func() { _ = db.Close() }()

	// Verify all tables exist by querying sqlite_master
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('issues', 'labels', 'dependencies', 'comments', 'blocked_issues_cache')`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 5, count, "expected 5 tables")

	// Verify ready_issues view exists
	err = db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='view' AND name='ready_issues'`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "expected ready_issues view")
}

func TestNewTestDB_TablesExist(t *testing.T) {
	db := NewTestDB(t)
	defer func() { _ = db.Close() }()

	// Test each table is queryable via COUNT
	tables := []string{"issues", "labels", "dependencies", "comments", "blocked_issues_cache"}
	for _, table := range tables {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
		require.NoError(t, err, "table %s should be queryable", table)
	}

	// Test view is queryable
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM ready_issues").Scan(&count)
	require.NoError(t, err, "view ready_issues should be queryable")
}

func TestNewClassicTestDB_HasNoGasTownColumns(t *testing.T) {
	db := NewClassicTestDB(t)
	defer func() { _ = db.Close() }()

	gasTownCols := []string{"hook_bead", "role_bead", "agent_state", "last_activity", "role_type", "rig", "mol_type"}
	rows, err := db.Query(`PRAGMA table_info(issues)`)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()

	var presentCols []string
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue any
		require.NoError(t, rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk))
		for _, g := range gasTownCols {
			if name == g {
				presentCols = append(presentCols, name)
			}
		}
	}
	require.NoError(t, rows.Err())
	require.Empty(t, presentCols, "classic schema must not contain GasTown columns: %v", presentCols)
}

func TestNewClassicBuilder_WithStandardTestData_Succeeds(t *testing.T) {
	db := NewClassicTestDB(t)
	defer func() { _ = db.Close() }()

	NewClassicBuilder(t, db).WithStandardTestData().Build()

	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM issues`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 6, count, "expected 6 issues from standard test data")
}

func TestNewBuilder_FullSchema_StillInsertsGasTownColumns(t *testing.T) {
	db := NewTestDB(t)
	defer func() { _ = db.Close() }()

	NewBuilder(t, db).WithIssue("i1", HookBead("hook-val"), RoleBead("role-val")).Build()

	var hookBead, roleBead string
	err := db.QueryRow(`SELECT hook_bead, role_bead FROM issues WHERE id = 'i1'`).Scan(&hookBead, &roleBead)
	require.NoError(t, err)
	require.Equal(t, "hook-val", hookBead)
	require.Equal(t, "role-val", roleBead)
}

func TestNewTestDB_IssueColumns(t *testing.T) {
	db := NewTestDB(t)
	defer func() { _ = db.Close() }()

	// Insert a test issue with all columns
	_, err := db.Exec(`INSERT INTO issues
		(id, title, description, design, acceptance_criteria, notes, status, priority, issue_type, assignee, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		"test-id", "Test Title", "Test Desc", "Test Design", "Test AC", "Test Notes", "open", 1, "task", "alice")
	require.NoError(t, err)

	// Verify all columns exist and are readable
	var id, title, desc, design, ac, notes, status, issueType string
	var priority int
	var assignee *string
	err = db.QueryRow(`SELECT id, title, description, design, acceptance_criteria, notes, status, priority, issue_type, assignee FROM issues WHERE id = ?`, "test-id").
		Scan(&id, &title, &desc, &design, &ac, &notes, &status, &priority, &issueType, &assignee)
	require.NoError(t, err)
	require.Equal(t, "test-id", id)
	require.Equal(t, "Test Title", title)
	require.Equal(t, "Test Desc", desc)
	require.Equal(t, "Test Design", design)
	require.Equal(t, "Test AC", ac)
	require.Equal(t, "Test Notes", notes)
	require.Equal(t, "open", status)
	require.Equal(t, 1, priority)
	require.Equal(t, "task", issueType)
	require.NotNil(t, assignee)
	require.Equal(t, "alice", *assignee)
}
