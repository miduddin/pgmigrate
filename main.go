package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type logger struct{}

func (l *logger) Printf(format string, args ...any) {
	fmt.Printf(format, args...)
}

func (l *logger) Verbose() bool {
	return false
}

func main() {
	if len(os.Args) < 3 {
		fmt.Print("First arg: pgservice name, second arg: migrations dir.")
		os.Exit(1)
	}

	svc, err := loadPgService(os.Args[1])
	panicIfError(err)
	if svc["search_path"] == "" {
		svc["search_path"] = "public"
	}

	db, err := sql.Open("postgres", fmt.Sprintf(
		"postgres://%s@%s/%s?sslmode=disable&search_path=%s",
		svc["user"], svc["host"], svc["dbname"], svc["search_path"],
	))
	panicIfError(err)
	panicIfError(db.Ping())

	dr, err := postgres.WithInstance(db, &postgres.Config{})
	panicIfError(err)

	wd, _ := os.Getwd()
	m, err := migrate.NewWithDatabaseInstance("file://"+filepath.Join(wd, os.Args[2]), "postgres", dr)
	panicIfError(err)
	m.Log = &logger{}

	ver, dirty, err := m.Version()
	panicIfError(err)

	fmt.Printf(
		"Connected to host=%s db=%s schema=%s user=%s\n",
		yellow(svc["host"]), yellow(svc["dbname"]), yellow(svc["search_path"]), yellow(svc["user"]),
	)
	fmt.Printf("Current migration version: %s (dirty: %v)\n\n", white("%d", ver), dirty)

	if dirty {
		fmt.Println(red("Fix dirty migration version first!"))
		os.Exit(1)
	}

	fmt.Printf("Pending migrations:\n\n")

	if !printPendingMigrations(ver, os.Args[2]) {
		fmt.Println(green("None, database is up to date!"))
		return
	}

	fmt.Print(yellow("Continue with migrations? (y/n): "))
	var in string
	fmt.Scanln(&in)
	if strings.ToLower(in) != "y" {
		fmt.Println("Aborting.")
		return
	}

	panicIfError(m.Up())

	fmt.Println(green("Done."))
}

func printPendingMigrations(version uint, path string) bool {
	files, err := os.ReadDir(path)
	panicIfError(err)

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	vs := fmt.Sprint(version)
	count := 0
	for _, f := range files {
		if f.Name() > vs && !strings.HasPrefix(f.Name(), vs) && strings.HasSuffix(f.Name(), ".up.sql") {
			b, err := os.ReadFile(filepath.Join(path, f.Name()))
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
