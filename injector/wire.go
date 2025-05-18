//go:build wireinject
// +build wireinject

package injector

import (
	"github.com/arifin2018/splitbill-arifin.git/controllers"
	splitbillcontollers "github.com/arifin2018/splitbill-arifin.git/controllers/SplitbillContollers"
	splitbillservices "github.com/arifin2018/splitbill-arifin.git/services/SplitbillServices"
	"github.com/google/wire"
)

var splitbilController = wire.NewSet(
	splitbillservices.NewSplitbillServiceImpl,
	wire.Bind(new(splitbillservices.SplibillService), new(*splitbillservices.SplibillServiceImpl)),
	splitbillcontollers.NewSplitbilController,
	wire.Bind(new(splitbillcontollers.SplitbilController), new(*splitbillcontollers.SplitbillControllerImpl)),
)

var setAllControllers = wire.NewSet(
	// ProvideDB,
	splitbilController,
	wire.Struct(new(controllers.AllControllers), "*"),
)

func InitializeController() *controllers.AllControllers {
	// wire.Build(setLoginController, controllers.NewAllControllers)
	wire.Build(setAllControllers)
	return &controllers.AllControllers{}
}
