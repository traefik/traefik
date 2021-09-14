package tcp

//func TestRouter(t *testing.T) {
//	router, err := NewRouter()
//	require.NoError(t, err)
//	assert.NotNil(t, router)
//
//	route := NewRoute(nil)
//	require.NotNil(t, route)
//
//	route.AddMatcher(NewClientIP("10.1"))
//	assert.Len(t, route.matchers, 1)
//	route.AddMatcher(NewSNIHost("foo"))
//	assert.Len(t, route.matchers, 2)
//}
//
//func TestAddMatchers(t *testing.T) {
//	route := NewRoute(nil)
//	require.NotNil(t, route)
//
//	route.AddMatcher(NewClientIP("10.1"))
//	assert.Len(t, route.matchers, 1)
//	route.AddMatcher(NewSNIHost("foo"))
//	assert.Len(t, route.matchers, 2)
//}
//
//func TestAddRoutes(t *testing.T) {
//	router, err := NewRouter()
//	require.NoError(t, err)
//	assert.NotNil(t, router)
//
//	route := NewRoute(nil)
//	require.NotNil(t, route)
//	route.AddMatcher(NewClientIP("10.1"))
//	router.AddRoute(route)
//
//	assert.Len(t, router.routes, 1)
//}
