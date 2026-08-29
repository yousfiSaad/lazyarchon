package styling

import (
	configpkg "github.com/yousfisaad/lazyarchon/v2/internal/shared/config"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/styling"
)

// Provider provides style context creation without dependency injection
type Provider struct {
	config interfaces.ConfigProvider
}

// NewProvider creates a new styling provider
func NewProvider(config interfaces.ConfigProvider) *Provider {
	return &Provider{
		config: config,
	}
}

// CreateStyleContext creates a basic style context with theme adapter
func (p *Provider) CreateStyleContext(forceBackground bool) *styling.StyleContext {
	// Create a basic style context with theme adapter
	themeAdapter := &styling.ThemeAdapter{
		TodoColor:     styling.CurrentTheme.TodoColor,
		DoingColor:    styling.CurrentTheme.DoingColor,
		ReviewColor:   styling.CurrentTheme.ReviewColor,
		DoneColor:     styling.CurrentTheme.DoneColor,
		HeaderColor:   styling.CurrentTheme.HeaderColor,
		MutedColor:    styling.CurrentTheme.MutedColor,
		AccentColor:   styling.CurrentTheme.AccentColor,
		StatusColor:   styling.CurrentTheme.StatusColor,
		FeatureColors: styling.CurrentTheme.FeatureColors,
		Name:          styling.CurrentTheme.Name,
	}
	return styling.NewStyleContext(themeAdapter, p.config)
}

// GetTheme returns the theme configuration
func (p *Provider) GetTheme() *configpkg.ThemeConfig {
	return p.config.GetTheme()
}
