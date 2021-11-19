package repository

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"testing"

	"gopkg.in/yaml.v2"
)

type FirestoreClientMock struct {}

func (f *FirestoreClientMock) GetItemsFromCollection(ctx context.Context, path string) (items []*map[string]interface{}, err error) {
	fmt.Println("GetItemsFromCollection")
	return games, nil
}
func (f *FirestoreClientMock) StoreItemToCollection(ctx context.Context, path string, item *map[string]interface{}) (id string, err error) {
	games = append(games, item)
	return "3", nil
}
func (f *FirestoreClientMock) UpdateItem(ctx context.Context, path string, refID string, updates map[string]interface{}) (err error) {
	return nil
}
func (f *FirestoreClientMock) UpdateValueOnItem(ctx context.Context, path string, refID string, key string, value interface{}) (err error) {
	return nil
}

//go:embed haiku_testdata.yaml
var testData []byte
var games []*map[string]interface{}

func TestMain(m *testing.M) {
	if err := yaml.Unmarshal(testData, &games); err != nil {
		fmt.Println("yaml ミスった")
		os.Exit(1)
	}
	for _, v := range games {
		fmt.Printf("%+v\n",v)
	}

	exitVal := m.Run()

	os.Exit(exitVal)
}

func TestSample(t *testing.T) {
	ctx := context.Background()
	f := &FirestoreClientMock{}
	{
		games, _ := f.GetItemsFromCollection(ctx, "make")
		for _, v := range games {
			fmt.Printf("%+v\n",v)
		}
	}
	repo, err := NewHaikuRepository(ctx, f)
	if err != nil {
		t.Fatalf("作れなかった: %+v", err)
	}

	game, poet := repo.GetGameAndPoetByPoetID("1")
	if game == nil || poet == nil {
		t.Fatal("nil")
	}
	if game.RefID != "1" {
		t.Logf("%v\n", game)
		t.Fatal("poet1 is not on game1")
	}
	if poet.NextPoemNumber != 1 {
		t.Fatal("poet1's next poem number is not 1")
	}

	poem := repo.GetNextPoem(poet, game)
	if poem.PoemRunes[0].Rune != "い" {
		t.Fatal("next poem is not 1")
	}

	poem.PoemRunes = append(poem.PoemRunes, NewPoemRune(poet.ID, "う"))

	poem2 := repo.GetNextPoem(poet, game)
	t.Logf("poem: %v, poem2: %v", poem.FormatHaiku(), poem2.FormatHaiku())
	if poem.FormatHaiku() != poem2.FormatHaiku() {
		t.Fatal("アドレス参照してるけど駄目っぽい")
	}
}
