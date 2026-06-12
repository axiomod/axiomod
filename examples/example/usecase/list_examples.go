package usecase

import (
	"context"

	"github.com/axiomod/axiomod/examples/example/repository"
)

// ListExamplesInput represents the input for listing examples
type ListExamplesInput struct {
	Name      string `json:"name"`
	ValueType string `json:"valueType"`
	Tag       string `json:"tag"`
	Limit     int    `json:"limit" validate:"gte=0"`
	Offset    int    `json:"offset" validate:"gte=0"`
}

// ListExamplesItem represents a single example in the list output
type ListExamplesItem struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	ValueType   string   `json:"valueType"`
	Count       int      `json:"count"`
	Tags        []string `json:"tags"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}

// ListExamplesOutput represents the output of listing examples
type ListExamplesOutput struct {
	Items []ListExamplesItem `json:"items"`
	Total int                `json:"total"`
}

// ListExamplesUseCase defines the use case for listing examples
type ListExamplesUseCase struct {
	repo repository.ExampleRepository
}

// NewListExamplesUseCase creates a new ListExamplesUseCase
func NewListExamplesUseCase(repo repository.ExampleRepository) *ListExamplesUseCase {
	return &ListExamplesUseCase{
		repo: repo,
	}
}

// Execute executes the use case
func (uc *ListExamplesUseCase) Execute(ctx context.Context, input ListExamplesInput) (*ListExamplesOutput, error) {
	filter := repository.ExampleFilter{
		Name:      input.Name,
		ValueType: input.ValueType,
		Tag:       input.Tag,
		Limit:     input.Limit,
		Offset:    input.Offset,
	}

	examples, err := uc.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]ListExamplesItem, 0, len(examples))
	for _, example := range examples {
		items = append(items, ListExamplesItem{
			ID:          example.ID,
			Name:        example.Name,
			Description: example.Description,
			ValueType:   example.Value.Type,
			Count:       example.Value.Count,
			Tags:        example.Value.Tags,
			CreatedAt:   example.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   example.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return &ListExamplesOutput{
		Items: items,
		Total: len(items),
	}, nil
}
