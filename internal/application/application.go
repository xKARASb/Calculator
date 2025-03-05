package application

import (
	"log"
	"net/http"

	"https://github.com/xKARASb/Calculator/internal/web"
	"https://github.com/xKARASb/Calculator/pkg/dir"
	"https://github.com/xKARASb/Calculator/pkg/rpn"

	"github.com/gorilla/mux"
)

var Expressions = make(map[IDExpression]*Expression)

var Tasks = rpn.NewConcurrentTaskMap()

type Application struct {
	Config       *config
	Agent        http.Client
	NumGoroutine int
	Router       *mux.Router
}

func New() *Application {
	return &Application{
		Router: mux.NewRouter(),
		Config: newConfig(),
	}
}

func (app *Application) RunServer() {
	rpn.InitEnv(dir.Env_file())
	startServer := make(chan struct{}, 1)
	go func() {
		startServer <- struct{}{}
		if app.Config.Debug {
			log.Println("Orkestrator Runned")
		}
		err := http.ListenAndServe(":8080", nil)
		panic(err)
	}()
	app.Router.HandleFunc("/api/v1/calculate", app.AddExpressionHandler)
	app.Router.HandleFunc("/api/v1/expressions/{id}", app.GetExpressionHandler)
	app.Router.HandleFunc("/api/v1/expressions", app.GetExpressionsHandler)
	app.Router.HandleFunc("/api/v1/internal/task", app.TaskHandler)
	if app.Config.Web {
		web.HandleToRouter(app.Router)
	}
	http.Handle("/", app.Router)
	<-startServer
	panic(app.runAgent())
}
