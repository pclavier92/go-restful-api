package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pclavier92/go-restful-api/pkg/errors"
	"github.com/pclavier92/go-restful-api/pkg/logs"
)

// Controller is a function which takes a context and returns an status code,
// an interface{} which will be converted to json and maybe an error.
type Controller func(*Context) (int, interface{}, error)

// HandlerFunc is a classic handler for gin
type HandlerFunc = gin.HandlerFunc

// Context has the information present in an HTTP Request, simple wrapper for gin.Context
type Context struct {
	*gin.Context
	ID string
}

// SetTest will activate gin's test mode
func SetTest() {
	gin.SetMode(gin.TestMode)
}

// Engine is a simple wrapper for gin.Engine
type Engine struct {
	gin  *gin.Engine
	port string
}

// New returns a new engine for you to use as you please
func New(port string) *Engine {
	g := gin.New()
	g.Use(gin.Recovery())
	g.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	return &Engine{g, port}
}

// Static will serve static content
func (e *Engine) Static(path, location string) {
	e.gin.Static(path, location)
}

// UseLogger will activate the default gin logger for this engine
func (e *Engine) UseLogger() {
	e.gin.Use(gin.Logger())
}

// ServeHTTP makes a request to the engine
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	e.gin.ServeHTTP(w, req)
}

// Group returns a RouterGroup so you can organize paths
func (e *Engine) Group(prefix string) *RouterGroup {
	return &RouterGroup{e.gin.Group(prefix)}
}

// GET registers a path to go to a controller
func (e *Engine) GET(path string, fn Controller) {
	e.gin.GET(path, adapt(fn))
}

// PATCH registers a path to go to a controller
func (e *Engine) PATCH(path string, fn Controller) {
	e.gin.PATCH(path, adapt(fn))
}

// POST registers a path to go to a controller
func (e *Engine) POST(path string, fn Controller) {
	e.gin.POST(path, adapt(fn))
}

// PUT registers a path to go to a controller
func (e *Engine) PUT(path string, fn Controller) {
	e.gin.PUT(path, adapt(fn))
}

// Any registers a path to go to a controller with any method
func (e *Engine) Any(path string, fn Controller) {
	e.gin.Any(path, adapt(fn))
}

// Run will start the engine! *waves flag*
func (e *Engine) Run() error {
	return e.gin.Run(":" + e.port)
}

// RouterGroup is a simple router where we can organize paths
type RouterGroup struct {
	gin *gin.RouterGroup
}

// UseLogger will activate the default gin logger for this engine
func (r *RouterGroup) UseLogger() {
	r.gin.Use(gin.Logger())
}

// GET registers a path to go to a controller
func (r *RouterGroup) GET(path string, fn Controller) {
	r.gin.GET(path, adapt(fn))
}

// POST registers a path to go to a controller
func (r *RouterGroup) POST(path string, fn Controller) {
	r.gin.POST(path, adapt(fn))
}

// PUT registers a path to go to a controller
func (r *RouterGroup) PUT(path string, fn Controller) {
	r.gin.PUT(path, adapt(fn))
}

// PATCH registers a path to go to a controller
func (r *RouterGroup) PATCH(path string, fn Controller) {
	r.gin.PATCH(path, adapt(fn))
}

// DELETE registers a path to go to a controller
func (r *RouterGroup) DELETE(path string, fn Controller) {
	r.gin.DELETE(path, adapt(fn))
}

// Group can create subgroups in a RouterGroup
func (r *RouterGroup) Group(prefix string) *RouterGroup {
	return &RouterGroup{r.gin.Group(prefix)}
}

// StatusJSON is used to communicate statuses to the user
type StatusJSON struct {
	Data statusData `json:"data"`
}
type statusData struct {
	ID    string      `json:"id"`
	Type  string      `json:"type"`
	Attrs statusAttrs `json:"attributes"`
}
type statusAttrs struct {
	Status string `json:"status"`
}

// MakeStatus returns a new StatusJSON to display to the user.
func MakeStatus(id, status string) StatusJSON {
	return StatusJSON{statusData{id, "status", statusAttrs{status}}}
}

// adapt converts from a function taking a context and returning
// an status and a json or a string or an apiErr.
// it will also set a random id to the context
func adapt(cr Controller) gin.HandlerFunc {
	type error struct {
		Error string `json:"error"`
	}
	return func(c *gin.Context) {
		// TODO: make some id random
		cc := Context{c, "someID"}
		code, ctx, err := cr(&cc)
		if err != nil {
			if e, ok := err.(*errors.Chain); ok {
				e.Pkg.Log.Info("Error in API", logs.I{
					"error":    err.Error(),
					"external": e.External,
					"code":     code,
				})
				e := error{e.External}
				c.JSON(code, e)
				return
			}
			c.JSON(code, error{"unkonwn error :("})
			return
		}
		if code == 302 {
			url := ctx.(string)
			c.Redirect(302, url)
			return
		}
		if str, ok := ctx.(string); ok {
			ctx = MakeStatus(str, "ok")
		}
		c.JSON(code, ctx)
	}
}
