package icons

import (
	"context"
	"io"

	swarmicons "github.com/frostybee/go-swarm-icons"
	"github.com/frostybee/go-swarm-icons/lucide"

	"github.com/a-h/templ"
)

var Manager = swarmicons.Default(
	"lucide",
	lucide.Provider(),
)

// Icon renders a go-swarm-icons icon as a templ component.
func Icon(name string, class string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		icon, err := Manager.Get(name, map[string]string{
			"class": class,
		})
		if err != nil {
			return err
		}

		_, err = io.WriteString(w, icon.ToHTML())
		return err
	})
}
