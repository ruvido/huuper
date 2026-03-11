package static

import (
	"os"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func Register(se *core.ServeEvent) {
	se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))
}
