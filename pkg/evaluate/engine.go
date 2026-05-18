package evaluate

import (
	"fmt"
	"log"
	"os"
	"sort"

	"k8strike/pkg/util"
)

type Context struct {
	Logger *log.Logger
}

func NewContext(logger *log.Logger) *Context {
	if logger == nil {
		logger = log.New(os.Stderr, "", log.LstdFlags)
	}
	return &Context{Logger: logger}
}

type CheckFunc func(*Context) error

type Check struct {
	ID          string
	Title       string
	Description string
	Run         CheckFunc
}

func (c Check) execute(ctx *Context) error {
	if c.Run == nil {
		return nil
	}
	return c.Run(ctx)
}

type Category struct {
	ID     string
	Title  string
	Checks []Check
}

func (c Category) run(ctx *Context) {
	util.PrintH2(c.Title)
	logger := loggerFromContext(ctx)
	for _, check := range c.Checks {
		if err := check.execute(ctx); err != nil {
			logger.Printf("check %s failed: %v", readableCheckLabel(check), err)
		}
	}
}

type Profile struct {
	ID         string
	Title      string
	Categories []Category
}

func (p Profile) run(ctx *Context) {
	for _, category := range p.Categories {
		category.run(ctx)
	}
}

type Evaluator struct {
	profiles map[string]Profile
}

func NewEvaluator() *Evaluator {
	e := &Evaluator{profiles: make(map[string]Profile)}
	for _, profile := range defaultProfiles() {
		e.RegisterProfile(profile)
	}
	return e
}

func (e *Evaluator) RegisterProfile(profile Profile) {
	if e.profiles == nil {
		e.profiles = make(map[string]Profile)
	}
	e.profiles[profile.ID] = profile
}

func (e *Evaluator) Profile(id string) (Profile, bool) {
	profile, ok := e.profiles[id]
	return profile, ok
}

func (e *Evaluator) Profiles() []Profile {
	out := make([]Profile, 0, len(e.profiles))
	for _, profile := range e.profiles {
		out = append(out, profile)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func (e *Evaluator) RunProfile(id string, ctx *Context) error {
	profile, ok := e.profiles[id]
	if !ok {
		return fmt.Errorf("unknown profile %q", id)
	}
	if ctx == nil {
		ctx = NewContext(nil)
	}
	profile.run(ctx)
	return nil
}

func loggerFromContext(ctx *Context) *log.Logger {
	if ctx != nil && ctx.Logger != nil {
		return ctx.Logger
	}
	return log.Default()
}

func readableCheckLabel(check Check) string {
	if check.ID != "" {
		return fmt.Sprintf("%s (%s)", check.Title, check.ID)
	}
	return check.Title
}
