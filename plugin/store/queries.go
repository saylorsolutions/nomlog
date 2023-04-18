package store

import "fmt"

const (
	createTable = `
create table if not exists %s (
	evt_id integer primary key
)`
)

func init() {
	fmt.Println()
}
