package instructor

import (
	anthropic "github.com/liushuangls/go-anthropic/v2"
)

type InstructorAnthropic struct {
	*anthropic.Client

	provider   Provider
	mode       Mode
	maxRetries int
	validate   bool
}

var _ Instructor = &InstructorAnthropic{}

func FromAnthropic(client *anthropic.Client, opts ...Options) *InstructorAnthropic {

	options := mergeOptions(opts...)

	i := &InstructorAnthropic{
		Client: client,

		provider:   ProviderOpenAI,
		mode:       *options.Mode,
		maxRetries: *options.MaxRetries,
		validate:   *options.validate,
	}
	return i
}

func (i *InstructorAnthropic) MaxRetries() int {
	return i.maxRetries
}

func (i *InstructorAnthropic) Mode() string {
	return i.mode
}

func (i *InstructorAnthropic) Provider() string {
	return i.provider
}
func (i *InstructorAnthropic) Validate() bool {
	return i.validate
}
