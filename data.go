package main

//go:generate go run ./scripts/generate_snapshot.go -- snapshot_data.go

type emojiRanking struct {
	char, id, name string
	score          int
}
