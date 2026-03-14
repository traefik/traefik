package snippet

import (
	"fmt"
	"slices"

	"github.com/tufanbarisyildirim/gonginx/config"
)

const (
	contextServer       = "server"
	contextLocation     = "location"
	contextIf           = "if"
	contextIfInLocation = "if_in_location"
)

var directiveContexts = map[string][]string{
	"add_header":               {contextServer, contextLocation, contextIfInLocation},
	"more_set_headers":         {contextServer, contextLocation, contextIf},
	"more_clear_headers":       {contextServer, contextLocation, contextIf},
	"proxy_set_header":         {contextServer, contextLocation},
	"more_set_input_headers":   {contextServer, contextLocation, contextIf},
	"more_clear_input_headers": {contextServer, contextLocation, contextIf},
	"if":                       {contextServer, contextLocation},
	"set":                      {contextServer, contextLocation, contextIf},
	"return":                   {contextServer, contextLocation, contextIf},
	"rewrite":                  {contextServer, contextLocation, contextIf},
	"location":                 {contextServer},
	"allow":                    {contextServer, contextLocation, contextIf},
	"deny":                     {contextServer, contextLocation, contextIf},
	"proxy_hide_header":        {contextServer, contextLocation},
	"expires":                  {contextServer, contextLocation, contextIfInLocation},
}

// isAllowedInContext checks if the directive is allowed in the context of its parent directive.
func isAllowedInContext(directive config.IDirective) error {
	ctx := directive.GetParent().GetName()
	allowedCtxs := directiveContexts[directive.GetName()]

	if slices.Contains(allowedCtxs, ctx) {
		return nil
	}

	if slices.Contains(allowedCtxs, contextIfInLocation) && ctx == contextIf {
		// Here we are checking if the parent of the "if" directive is a "location" directive,
		// which means that the "if" directive is inside a "location" block.
		if directive.GetParent().GetParent().GetName() == contextLocation {
			return nil
		}
	}

	return fmt.Errorf("context %s is not valid for this directive %s: %+v ", ctx, directive.GetName(), allowedCtxs)
}
