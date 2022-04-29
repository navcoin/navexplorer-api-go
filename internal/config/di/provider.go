// provider.go
package di

import (
	"github.com/sarulabs/dingo/v4"
)

// Redefine your own provider by overriding the Load method of the dingo.BaseProvider.
type Provider struct {
	dingo.BaseProvider
}

func (p *Provider) Load() error {
	if err := p.AddDefSlice(Definitions); err != nil {
		return err
	}

	return nil
}
