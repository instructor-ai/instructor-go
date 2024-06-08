package instructor

import (
	cohere "github.com/cohere-ai/cohere-go/v2/client"
)

type InstructorCohere struct {
	*cohere.Client

	provider      Provider
	mode          Mode
	maxRetries    int
	withValidator bool
}

var _ Instructor = &InstructorCohere{}

func FromCohere(client *cohere.Client, opts ...Options) *InstructorCohere {

	options := mergeOptions(opts...)

	i := &InstructorCohere{
		Client: client,

		provider:   ProviderCohere,
		mode:       *options.Mode,
		maxRetries: *options.MaxRetries,
	}
	return i
}

func (i *InstructorCohere) Provider() string {
	return i.provider
}

func (i *InstructorCohere) Mode() string {
	return i.mode
}

func (i *InstructorCohere) MaxRetries() int {
	return i.maxRetries
}
func (i *InstructorCohere) WithValidator() bool {
	return i.withValidator
}
