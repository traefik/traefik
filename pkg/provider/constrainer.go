package provider

import "github.com/containous/traefik/pkg/types"

// Constrainer Filter services by constraint, matching with Traefik tags.
type Constrainer struct {
	Constraints []*types.Constraint `description:"Filter services by constraint, matching with Traefik tags." export:"true"`
}

// MatchConstraints must match with EVERY single constraint
// returns first constraint that do not match or nil.
func (c *Constrainer) MatchConstraints(tags []string) (bool, *types.Constraint) {
	// if there is no tags and no constraints, filtering is disabled
	if len(tags) == 0 && len(c.Constraints) == 0 {
		return true, nil
	}

	for _, constraint := range c.Constraints {
		// xor: if ok and constraint.MustMatch are equal, then no tag is currently matching with the constraint
		if ok := constraint.MatchConstraintWithAtLeastOneTag(tags); ok != constraint.MustMatch {
			return false, constraint
		}
	}

	// If no constraint or every constraints matching
	return true, nil
}
