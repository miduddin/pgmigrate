package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage:\n\tpgmigrate <pgservice> <migrations_dir>")
		os.Exit(1)
	}

	db, err := pgConnect(os.Args[1])
	panicIfError(err)

	m, err := newMigrate(db, os.Args[2])
	panicIfError(err)

	ver, dirty, err := currentMigrationVersion(m)
	panicIfError(err)
	if dirty {
		fmt.Println(red("Fix dirty migration version first!"))
		os.Exit(1)
	}

	err = up(m, *ver, os.Args[2])
	panicIfError(err)

	fmt.Println(green("Done!"))
}

type logger struct{}

func (l *logger) Printf(format string, args ...any) {
	fmt.Printf(format, args...)
}

func (l *logger) Verbose() bool {
	return false
}

func newMigrate(db *sql.DB, migrationDir string) (*migrate.Migrate, error) {
	dr, _ := postgres.WithInstance(db, &postgres.Config{})
	wd, _ := os.Getwd()
	m, err := migrate.NewWithDatabaseInstance("file://"+filepath.Join(wd, migrationDir), "postgres", dr)
	if err != nil {
		return nil, fmt.Errorf("init migrate: %w", err)
	}

	m.Log = &logger{}
	return m, nil
}

func currentMigrationVersion(m *migrate.Migrate) (*uint, bool, error) {
	ver, dirty, err := m.Version()
	if err != nil {
		if err.Error() == "no migration" {
			fmt.Println("First time running migration.")
		} else {
			return nil, false, fmt.Errorf("get current migration version: %w", err)
		}
	} else {
		fmt.Printf("Current migration version: %s (dirty: %v)\n\n", white("%d", ver), dirty)
	}

	return &ver, dirty, nil
}

func up(m *migrate.Migrate, ver uint, migrationDir string) error {
	fmt.Printf("Pending migrations:\n\n")
	if !printPendingMigrations(ver, migrationDir) {
		fmt.Println(green("None, database is up to date!"))
		return nil
	}

	fmt.Print(yellow("Continue with migrations? (y/n): "))
	var in string
	_, _ = fmt.Scanln(&in)
	if strings.ToLower(in) != "y" {
		fmt.Println("Aborting.")
		return nil
	}

	return m.Up()
}

func printPendingMigrations(currentVersion uint, migrationDir string) bool {
	files, err := os.ReadDir(migrationDir)
	panicIfError(err)

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	count := 0
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".up.sql") {
			continue
		}

		fvs := strings.Split(f.Name(), "_")[0]
		fv, _ := strconv.ParseUint(fvs, 10, 64)
		if uint(fv) > currentVersion {
			b, err := os.ReadFile(filepath.Join(migrationDir, f.Name()))
			panicIfError(err)

			count++
			fmt.Print(white("%d. %s\n\n", count, f.Name()))
			fmt.Printf("\t%s\n\n", strings.ReplaceAll(string(b), "\n", "\n\t"))
		}
	}

	return count > 0
}

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}
