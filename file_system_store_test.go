package mond

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func createTempFile(t testing.TB, initialData string) (*os.File, func()) {
	t.Helper()

	tmpfile, err := ioutil.TempFile("", "db")

	if err != nil {
		t.Fatalf("could not create temp file %v", err)
	}

	tmpfile.Write([]byte(initialData))

	removeFile := func() {
		os.Remove(tmpfile.Name())
	}

	return tmpfile, removeFile
}

func TestFileSystemStore(t *testing.T) {
	t.Run("get apps and all logs of App1", func(t *testing.T) {
		database, cleanDatabase := createTempFile(t, `[
			{"app":"App1","logs":[
				{"timestamp":0,"ip":"","path":"","remoteIp":"","status":"","raw":"Test1"},
				{"timestamp":0,"ip":"","path":"","remoteIp":"","status":"","raw":"Test2"}
			]},
			{"app":"App2","logs":[
				{"timestamp":0,"ip":"","path":"","remoteIp":"","status":"","raw":"Test3"},
				{"timestamp":0,"ip":"","path":"","remoteIp":"","status":"","raw":"Test4"}
			]}
		]`)
		defer cleanDatabase()

		store, err := NewFileSystemAppsStore(database)

		assertNoError(t, err)

		got := store.GetAppNames()
		want := []string{
			"App1",
			"App2",
		}
		assertAppNamesEquals(t, got, want)

		gotLogs := store.GetAccessLogs("App1")
		wantLogs := AccessLogs{
			{Raw: "Test1"},
			{Raw: "Test2"},
		}
		assertAccessLogsEquals(t, gotLogs, wantLogs)
	})

	t.Run("store log for existing apps", func(t *testing.T) {
		database, cleanDatabase := createTempFile(t, `[
			{"app":"App1","logs":[
				{"timestamp":0,"ip":"","path":"","remoteIp":"","status":"","raw":"Test1"},
				{"timestamp":0,"ip":"","path":"","remoteIp":"","status":"","raw":"Test2"}
			]}
			]`)
		defer cleanDatabase()

		store, err := NewFileSystemAppsStore(database)
		assertNoError(t, err)

		store.RecordAccessLog("App1", AccessLog{Raw: "Test3"})

		got := store.GetAccessLogs("App1")
		want := AccessLogs{
			{Raw: "Test1"},
			{Raw: "Test2"},
			{Raw: "Test3"},
		}
		assertAccessLogsEquals(t, got, want)
	})

	t.Run("store logs for non existing apps", func(t *testing.T) {
		database, cleanDatabase := createTempFile(t, `[]`)
		defer cleanDatabase()

		store, err := NewFileSystemAppsStore(database)
		assertNoError(t, err)

		store.RecordAccessLog("App1", AccessLog{Raw: "Test"})

		got := store.GetAccessLogs("App1")
		want := AccessLogs{
			{Raw: "Test"},
		}
		assertAccessLogsEquals(t, got, want)
	})

	t.Run("works with an empty file", func(t *testing.T) {
		database, cleanDatabase := createTempFile(t, "")
		defer cleanDatabase()

		_, err := NewFileSystemAppsStore(database)

		assertNoError(t, err)
	})
}

func assertAppNamesEquals(t testing.TB, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func assertAccessLogsEquals(t testing.TB, got, want AccessLogs) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func assertNoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("didn't expect an error but got one, %v", err)
	}
}
