package prompttypes

import (
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
)

type Store struct {
	promptTypes map[string]*models.PromptType
}

func newStore(files []*promptTypeFiles) *Store {
	store := &Store{
		promptTypes: make(map[string]*models.PromptType),
	}

	for _, f := range files {
		store.promptTypes[f.name] = f.PromptType()
	}
	return store
}

func (s *Store) PromptTypes() (out []*models.PromptType) {
	for _, f := range s.promptTypes {
		out = append(out, f)
	}
	return out
}

func (s *Store) MustGetPromptType(name string) *models.PromptType {
	a, err := s.GetPromptType(name)
	if err != nil {
		panic(err)
	}
	return a
}

func (s *Store) GetPromptType(name string) (*models.PromptType, error) {
	a, found := s.promptTypes[name]
	if !found {
		return nil, datastore.NotFound
	}
	return a, nil
}
