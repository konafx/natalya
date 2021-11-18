package repository

import (
	"context"
	"fmt"
	"math"

	"github.com/konafx/natalya/service"
	"github.com/konafx/natalya/util"
)

type PoemRune struct {
	PoetID	string
	Rune	string
}

type Poem struct {
	PoemRunes []*PoemRune
}

func (poem *Poem) FormatHaiku() (ku string) {
	for k, v := range poem.PoemRunes {
		var x string
		switch k {
		case 5, 12:
			x = "　" + v.Rune
		default:
			x = v.Rune
		}
		ku = ku + x
	}
	return
}

type Poet struct {
	ID	string
	NextPoemNumber	uint
}

type HaikuGame struct {
	RefID	string
	Status	GameStatus
	Stage	uint
	NumberOfPoets	uint
	Poets	[]*Poet
	Poems	[]*Poem
}

type HaikuRepository struct {
	service	service.FirestoreClient
	games	[]*HaikuGame
}

func NewHaikuRepository(ctx context.Context, firestore service.FirestoreClient) (repo *HaikuRepository, err error) {
	repo = &HaikuRepository{}
	repo.service = firestore
	fmt.Println("get")
	games, err := repo.GetAllGames(ctx)
	if err != nil {
		return nil, err
	}
	repo.games = games
	return repo, nil
}

func NewHaikuGame(poets []*Poet) *HaikuGame {
	numberOfPoets := uint(len(poets))
	poems := make([]*Poem, numberOfPoets)
	return &HaikuGame{
		Stage:	1,
		Status:	GameStatusStart,
		NumberOfPoets:	numberOfPoets,
		Poems:	poems,
		Poets:	poets,
	}
}

func NewPoemRune(poetID string, _rune string) *PoemRune {
	return &PoemRune{poetID, _rune}
}

func (repo *HaikuRepository) GetAllGames(ctx context.Context) (games []*HaikuGame, err error) {
	fmt.Printf("%v", games)
	maps, err := repo.service.GetItemsFromCollection(ctx, "Haiku")
	if err != nil {
		return nil, err
	}
	fmt.Println("get!")
	for _, v := range maps {
		var game HaikuGame
		if ok, err := util.MapToStruct(*v, &game); !ok || err != nil {
			return nil, err
		}
		fmt.Printf("map: %+v\ngame:%+v\n\tpoet[0]: %+v, poem:%+v\n", v, game, *game.Poets[0], *game.Poems[0])
		games = append(games, &game)
	}
	return games, nil
}

// StoreGame 永続ストアへの保存を行い、成功後、オンメモリにも反映
func (repo *HaikuRepository) StoreGame(ctx context.Context, game *HaikuGame) error {
	var item map[string]interface{}
	if err := util.StructToMap(game, &item); err != nil {
		return err
	}
	RefID, err := repo.service.StoreItemToCollection(ctx, "Haiku", &item)
	if err != nil {
		return err
	}
	game.RefID = RefID
	repo.games = append(repo.games, game)
	return nil
}

func (repo *HaikuRepository) UpdateGame(ctx context.Context, game *HaikuGame) error {
	var updates map[string]interface{}
	if err := util.StructToMap(game, &updates); err != nil {
		return err
	}
	err := repo.service.UpdateItem(ctx, "Haiku", game.RefID, updates)
	if err != nil {
		return err
	}
	return nil
}

func (repo *HaikuRepository) GetGameByRefID(ctx context.Context, refID string) *HaikuGame {
	for _, v := range repo.games {
		if v.RefID == refID {
			return v
		}
	}
	return nil
}

func (repo *HaikuRepository) GetGameAndPoetByPoetID(poetID string) (game *HaikuGame, poet *Poet) {
	for _, v := range repo.games {
		poet := repo.GetPoetOnGameByID(poetID, v)
		if poet != nil {
			return v, poet
		}
	}
	return nil, nil
}

func (repo *HaikuRepository) GetNextPoem(poet *Poet, game *HaikuGame) *Poem {
	number := uint(math.Mod(float64(poet.NextPoemNumber), float64(game.NumberOfPoets)))
	poem := game.Poems[number]
	return poem
}

func (repo *HaikuRepository) GetPoetOnGameByID(poetID string, game *HaikuGame) *Poet {
	for _, v := range game.Poets {
		if v.ID == poetID {
			return v
		}
	}
	return nil
}
