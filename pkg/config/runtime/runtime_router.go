package runtime

import (
	"fmt"
	"slices"
	"strings"
)

// RouterGraph manages router tree relationships and validates circular dependencies
type RouterGraph struct {
	nodes map[string]*RouterNode
}

// RouterNode represents a router in the dependency graph
type RouterNode struct {
	router   *RouterInfo
	parents  []*RouterNode
	children []*RouterNode
}

// NewRouterGraph creates a new router graph
func NewRouterGraph(routers map[string]*RouterInfo) *RouterGraph {
	rg := &RouterGraph{
		nodes: make(map[string]*RouterNode),
	}

	for name, router := range routers {
		rg.addRouter(name, router)
	}

	rg.checkCircularDependencies()
	rg.checkOrphans()

	return rg
}

// addRouter adds a router to the graph and builds relationships
func (rg *RouterGraph) addRouter(name string, router *RouterInfo) error {
	node := &RouterNode{
		router:   router,
		parents:  make([]*RouterNode, 0),
		children: make([]*RouterNode, 0),
	}

	// Store node in graph
	rg.nodes[name] = node

	// Build parent relationships and set up children references
	if router.ParentRefs != nil {
		for _, parentName := range router.ParentRefs {
			if parentNode, exists := rg.nodes[parentName]; exists {
				// Add parent reference to this node
				node.parents = append(node.parents, parentNode)
				// Add this node as child to parent
				parentNode.children = append(parentNode.children, node)
				// Update parent's ChildRouterRefs field
				parentNode.router.ChildRouterRefs = append(parentNode.router.ChildRouterRefs, name)
			}
			// If parent doesn't exist yet, we'll handle it when it's added later
		}
	}

	// Update existing nodes that reference this router as parent
	for existingName, existingNode := range rg.nodes {
		if existingNode.router.ParentRefs != nil {
			for _, parentRef := range existingNode.router.ParentRefs {
				if parentRef == name {
					// This existing node references the new router as parent
					existingNode.parents = append(existingNode.parents, node)
					node.children = append(node.children, existingNode)
					// Update this router's ChildRouterRefs field
					router.ChildRouterRefs = append(router.ChildRouterRefs, existingName)
				}
			}
		}
	}

	return nil
}

func (rg *RouterGraph) checkCircularDependencies() {
	// Use DFS to detect cycles
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for nodeName := range rg.nodes {
		if !visited[nodeName] {
			if cycle := rg.dfsDetectCycle(nodeName, visited, recStack, []string{}); cycle != nil {
				// Format the cycle path for error message
				cyclePath := strings.Join(cycle, " -> ")
				cycleError := fmt.Errorf("circular dependency detected: %s", cyclePath)

				// Add error to all routers involved in the cycle
				for _, routerName := range cycle {
					if node, exists := rg.nodes[routerName]; exists {
						node.router.AddError(cycleError, true) // true = critical error
					}
				}
				return // Exit after first cycle detection
			}
		}
	}
}

func (rg *RouterGraph) checkOrphans() {
	for routerName, node := range rg.nodes {
		if node.router.ParentRefs == nil {
			continue
		}

		var invalidParents []string
		for _, parentRef := range node.router.ParentRefs {
			if _, exists := rg.nodes[parentRef]; !exists {
				invalidParents = append(invalidParents, parentRef)
			}
		}

		if len(invalidParents) > 0 {
			orphanError := fmt.Errorf("router %q references non-existent parent routers: %s",
				routerName, strings.Join(invalidParents, ", "))
			node.router.AddError(orphanError, true) // true = critical error
		}
	}
}

func (rg *RouterGraph) dfsDetectCycle(nodeName string, visited, recStack map[string]bool, path []string) []string {
	visited[nodeName] = true
	recStack[nodeName] = true
	path = append(path, nodeName)

	node := rg.nodes[nodeName]

	// Visit all parent nodes (dependencies)
	for _, parentNode := range node.parents {
		// Find parent name
		var parentName string
		for name, n := range rg.nodes {
			if n == parentNode {
				parentName = name
				break
			}
		}

		if !visited[parentName] {
			// Recursively visit parent
			if cycle := rg.dfsDetectCycle(parentName, visited, recStack, path); cycle != nil {
				return cycle
			}
		} else if recStack[parentName] {
			// Found a back edge - cycle detected
			// Find where the cycle starts in the path
			cycleStartIdx := slices.Index(path, parentName)
			if cycleStartIdx >= 0 {
				// Return the cycle path + the repeated node to show the loop
				cyclePath := append(path[cycleStartIdx:], parentName)
				return cyclePath
			}
		}
	}

	recStack[nodeName] = false
	return nil
}

// ProcessRouters processes a map of routers to populate ChildRouterRefs and detect validation errors.
// This function provides a simplified functional API alternative to the RouterGraph approach.
// It modifies the RouterInfo objects in-place to:
// 1. Populate ChildRouterRefs field for each router
// 2. Add error messages for circular dependencies
// 3. Add error messages for orphaned parent references
func ProcessRouters(routers map[string]*RouterInfo) map[string]*RouterInfo {
	if routers == nil {
		return nil
	}

	// Phase 1: Build parent-child relationships and populate ChildRouterRefs
	buildChildRelationships(routers)

	// Phase 2: Detect and report orphaned router references
	detectOrphans(routers)

	// Phase 3: Detect and report circular dependencies
	detectCycles(routers)

	return routers
}

// buildChildRelationships populates the ChildRouterRefs field for each router
// based on which routers reference them as parents
func buildChildRelationships(routers map[string]*RouterInfo) {
	// Clear existing ChildRouterRefs first
	for _, router := range routers {
		router.ChildRouterRefs = nil
	}

	// Build parent-to-children mapping
	parentToChildren := make(map[string][]string)

	for routerName, router := range routers {
		if router.ParentRefs != nil {
			for _, parentRef := range router.ParentRefs {
				parentToChildren[parentRef] = append(parentToChildren[parentRef], routerName)
			}
		}
	}

	// Populate ChildRouterRefs field for each parent router
	for parentName, childNames := range parentToChildren {
		if parent, exists := routers[parentName]; exists {
			parent.ChildRouterRefs = childNames
		}
	}
}

// detectOrphans finds routers that reference non-existent parent routers
// and adds error messages to those routers
func detectOrphans(routers map[string]*RouterInfo) {
	for routerName, router := range routers {
		if router.ParentRefs == nil {
			continue
		}

		var invalidParents []string
		for _, parentRef := range router.ParentRefs {
			if _, exists := routers[parentRef]; !exists {
				invalidParents = append(invalidParents, parentRef)
			}
		}

		if len(invalidParents) > 0 {
			orphanError := fmt.Errorf("router %q references non-existent parent routers: %s",
				routerName, strings.Join(invalidParents, ", "))
			router.AddError(orphanError, true) // true = critical error (disables router)
		}
	}
}

// detectCycles uses DFS to find circular dependencies between routers
// and adds error messages to all routers involved in any detected cycles
func detectCycles(routers map[string]*RouterInfo) {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for routerName := range routers {
		if !visited[routerName] {
			if cycle := dfsFindCycle(routerName, routers, visited, recStack, []string{}); cycle != nil {
				// Format the cycle path for error message
				cyclePath := strings.Join(cycle, " -> ")
				cycleError := fmt.Errorf("circular dependency detected: %s", cyclePath)

				// Add error to all routers involved in the cycle
				for _, routerName := range cycle {
					if router, exists := routers[routerName]; exists {
						router.AddError(cycleError, true) // true = critical error (disables router)
					}
				}
				return // Exit after first cycle detection
			}
		}
	}
}

// dfsFindCycle performs depth-first search to detect circular dependencies
// Returns the cycle path if found, nil otherwise
func dfsFindCycle(routerName string, routers map[string]*RouterInfo, visited, recStack map[string]bool, path []string) []string {
	visited[routerName] = true
	recStack[routerName] = true
	path = append(path, routerName)

	router, exists := routers[routerName]
	if !exists || router.ParentRefs == nil {
		recStack[routerName] = false
		return nil
	}

	for _, parentRef := range router.ParentRefs {
		if !visited[parentRef] {
			if cycle := dfsFindCycle(parentRef, routers, visited, recStack, path); cycle != nil {
				return cycle
			}
		} else if recStack[parentRef] {
			// Found a cycle - return the cycle path
			cycleStart := -1
			for i, node := range path {
				if node == parentRef {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cycle := make([]string, len(path)-cycleStart+1)
				copy(cycle, path[cycleStart:])
				cycle[len(cycle)-1] = parentRef // Close the cycle
				return cycle
			}
		}
	}

	recStack[routerName] = false
	return nil
}
