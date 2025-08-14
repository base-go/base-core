package router

// node represents a node in the routing tree
type node struct {
	path      string
	indices   string
	wildChild bool
	nType     nodeType
	priority  uint32
	children  []*node
	handler   HandlerFunc
}

type nodeType uint8

const (
	static nodeType = iota
	root
	param
	catchAll
)

// addRoute adds a route to the tree
func (n *node) addRoute(path string, handler HandlerFunc) {
	fullPath := path
	n.priority++

	// Empty tree
	if n.path == "" && n.children == nil {
		n.insertChild(path, fullPath, handler)
		n.nType = root
		return
	}

walk:
	for {
		// Find the longest common prefix
		i := longestCommonPrefix(path, n.path)

		// Split edge
		if i < len(n.path) {
			child := node{
				path:      n.path[i:],
				wildChild: n.wildChild,
				nType:     static,
				indices:   n.indices,
				children:  n.children,
				handler:   n.handler,
				priority:  n.priority - 1,
			}

			n.children = []*node{&child}
			n.indices = string([]byte{n.path[i]})
			n.path = path[:i]
			n.handler = nil
			n.wildChild = false
		}

		// Make new node a child of this node
		if i < len(path) {
			path = path[i:]

			if n.wildChild {
				n = n.children[len(n.children)-1]
				n.priority++

				// Check if the wildcard matches
				if len(path) >= len(n.path) && n.path == path[:len(n.path)] &&
					(len(n.path) >= len(path) || path[len(n.path)] == '/') {
					continue walk
				} else {
					panic("path segment '" + path +
						"' conflicts with existing wildcard '" + n.path +
						"' in path '" + fullPath + "'")
				}
			}

			idxc := path[0]

			// '/' after param
			if n.nType == param && idxc == '/' && len(n.children) == 1 {
				n = n.children[0]
				n.priority++
				continue walk
			}

			// Check if a child with the next path byte exists
			for i, c := range []byte(n.indices) {
				if c == idxc {
					n = n.children[i]
					n.priority++
					continue walk
				}
			}

			// Otherwise insert it
			if idxc != ':' && idxc != '*' {
				// []byte for proper unicode char conversion
				n.indices += string([]byte{idxc})
				child := &node{}
				n.children = append(n.children, child)
				n = child
			}
			n.insertChild(path, fullPath, handler)
			return
		}

		// Otherwise add handle to current node
		if n.handler != nil {
			panic("a handler is already registered for path '" + fullPath + "'")
		}
		n.handler = handler
		return
	}
}

// insertChild inserts a new child node
func (n *node) insertChild(path, fullPath string, handler HandlerFunc) {
	for {
		// Find wildcard segment
		wildcard, i, valid := findWildcard(path)
		if i < 0 {
			break
		}

		// The wildcard name must not contain ':' and '*'
		if !valid {
			panic("only one wildcard per path segment is allowed, has: '" +
				wildcard + "' in path '" + fullPath + "'")
		}

		// Check if the wildcard has a name
		if len(wildcard) < 2 {
			panic("wildcards must be named with a non-empty name in path '" + fullPath + "'")
		}

		// Check if this node has existing children which would be unreachable
		// Allow wildcard routes alongside static children if the wildcard comes after static routes
		if len(n.children) > 0 && wildcard[0] == '*' {
			// Catch-all wildcards (*) conflict with existing children
			panic("catch-all segment '" + wildcard +
				"' conflicts with existing children in path '" + fullPath + "'")
		}

		if wildcard[0] == ':' { // param
			if i > 0 {
				// Insert prefix before the current wildcard
				n.path = path[:i]
				path = path[i:]
			}

			n.wildChild = true
			child := &node{
				nType: param,
				path:  wildcard,
			}
			n.children = append(n.children, child)
			n = child

			// If the path doesn't end with the wildcard, then there will be
			// another non-wildcard subpath starting with '/'
			if len(wildcard) < len(path) {
				path = path[len(wildcard):]
				child := &node{
					priority: 1,
				}
				n.children = []*node{child}
				n = child
				continue
			}

			// Otherwise we're done. Insert the handler
			n.handler = handler
			return

		} else { // catchAll
			if i+len(wildcard) != len(path) {
				panic("catch-all routes are only allowed at the end of the path in path '" + fullPath + "'")
			}

			if len(n.path) > 0 && n.path[len(n.path)-1] == '/' {
				panic("catch-all conflicts with existing handler for the path segment root in path '" + fullPath + "'")
			}

			// Currently fixed width 1 for '/'
			i--
			if path[i] != '/' {
				panic("no / before catch-all in path '" + fullPath + "'")
			}

			n.path = path[:i]

			// First node: catchAll node with empty path
			child := &node{
				wildChild: true,
				nType:     catchAll,
			}
			n.children = []*node{child}
			n.indices = string('/')
			n = child

			// Second node: node holding the variable
			child = &node{
				path:     path[i:],
				nType:    catchAll,
				handler:  handler,
				priority: 1,
			}
			n.children = []*node{child}

			return
		}
	}

	// If no wildcard was found, simply insert the path and handler
	n.path = path
	n.handler = handler
}

// getValue returns the handler for the given path
func (n *node) getValue(path string) (handler HandlerFunc, params Params, tsr bool) {
walk: // Outer loop for walking the tree
	for {
		prefix := n.path
		if len(path) > len(prefix) {
			if path[:len(prefix)] == prefix {
				path = path[len(prefix):]

				// Always try static routes first, even if wildcard child exists
				idxc := path[0]
				for i, c := range []byte(n.indices) {
					if c == idxc {
						n = n.children[i]
						continue walk
					}
				}

				// If no static route found and wildcard child exists, try wildcard
				if n.wildChild {
					// Handle wildcard child (now always the last child)
					n = n.children[len(n.children)-1]
				} else {
					// Nothing found
					tsr = (path == "/" && n.handler != nil)
					return
				}
				switch n.nType {
				case param:
					// Find param end
					end := 0
					for end < len(path) && path[end] != '/' {
						end++
					}

					// Save param value
					if params == nil {
						params = make(Params, 0, 4)
					}
					params = append(params, Param{
						Key:   n.path[1:],
						Value: path[:end],
					})

					// We need to go deeper!
					if end < len(path) {
						if len(n.children) > 0 {
							path = path[end:]
							n = n.children[0]
							continue walk
						}

						// ... but we can't
						tsr = (len(path) == end+1)
						return
					}

					if handler = n.handler; handler != nil {
						return
					}
					if len(n.children) == 1 {
						// No handler found. Check if a handler for this path + a
						// trailing slash exists for TSR recommendation
						n = n.children[0]
						tsr = (n.path == "/" && n.handler != nil) || (n.path == "" && n.indices == "/")
					}
					return

				case catchAll:
					// Save param value
					if params == nil {
						params = make(Params, 0, 4)
					}
					params = append(params, Param{
						Key:   n.path[2:],
						Value: path,
					})

					handler = n.handler
					return

				default:
					panic("invalid node type")
				}
			}
		} else if path == prefix {
			// We should have reached the node containing the handler
			if handler = n.handler; handler != nil {
				return
			}

			// If there is no handler for this route, but this route has a
			// wildcard child, there must be a handler for this path with an
			// additional trailing slash
			if path == "/" && n.wildChild && n.nType != root {
				tsr = true
				return
			}

			// No handler found. Check if a handler for this path + a
			// trailing slash exists for trailing slash recommendation
			for i, index := range []byte(n.indices) {
				if index == '/' {
					n = n.children[i]
					tsr = (len(n.path) == 1 && n.handler != nil) ||
						(n.nType == catchAll && n.children[0].handler != nil)
					return
				}
			}

			return
		}

		// Nothing found. We can recommend to redirect to the same URL
		// without a trailing slash if a leaf exists for that path
		tsr = (path == "/") ||
			(len(prefix) == len(path)+1 && prefix[len(path)] == '/' &&
				path == prefix[:len(prefix)-1] && n.handler != nil)
		return
	}
}

// findWildcard finds a wildcard segment in the path
func findWildcard(path string) (wildcard string, i int, valid bool) {
	// Find start
	for start, c := range []byte(path) {
		// A wildcard starts with ':' (param) or '*' (catch-all)
		if c != ':' && c != '*' {
			continue
		}

		// Find end and check for invalid characters
		valid = true
		for end, c := range []byte(path[start+1:]) {
			switch c {
			case '/':
				return path[start : start+1+end], start, valid
			case ':', '*':
				valid = false
			}
		}
		return path[start:], start, valid
	}
	return "", -1, false
}

// longestCommonPrefix finds the longest common prefix
func longestCommonPrefix(a, b string) int {
	i := 0
	max := min(len(a), len(b))
	for i < max && a[i] == b[i] {
		i++
	}
	return i
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
