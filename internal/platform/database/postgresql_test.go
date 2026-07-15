package database

import "testing"

func TestNewPostgreSQLConnection(t *testing.T) {
	databaseConnection, err := NewPostgreSQLConnection("://invalid")
	if err == nil {
		t.Fatalf("NewPostgreSQLConnection() = %#v, %v", databaseConnection, err)
	}
}
