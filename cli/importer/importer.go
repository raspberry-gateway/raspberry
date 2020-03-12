package importer

import (
	"errors"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type APIImporterSource string

const (
	cmdName = "import"
	cmdDesc = "Imports a BluePrint/Swagger/WSDL file"
)

var (
	imp            *Importer
	errUnknownMode = errors.New("Unknown mode")
)

// Importer wraps the import functionlity
type Importer struct {
	input          *string
	swaggerMode    *bool
	bluePrintMode  *bool
	wsdlMode       *bool
	portNames      *string
	createAPI      *bool
	orgID          *string
	upstreamTarget *string
	asMock         *bool
	forAPI         *string
	asVersion      *string
}

func init() {
	imp = &Importer{}
}

// AddTo initializes an importer object.
func AddTo(app *kingpin.Application) {
	cmd := app.Command(cmdName, cmdDesc)
	imp.input = cmd.Arg("input file", "e.g. blueprint.json, swagger.json, service.wsdl etc.").String()
	imp.swaggerMode = cmd.Flag("swagger", "Use Swagger mode").Bool()
	imp.bluePrintMode = cmd.Flag("blueprint", "Use BluePrint mode").Bool()
	imp.wsdlMode = cmd.Flag("wsdl", "Use WSDL mode").Bool()
	imp.portNames = cmd.Flag("port-names", "Specify port name of each service in the WSDL file. Input format is comma separated list of serviceName:portName").String()
	imp.createAPI = cmd.Flag("create-api", "Creates a new API definition from the blueprint").Bool()
	imp.orgID = cmd.Flag("org-id", "assign the API Definition to this org_id (required with create-api)").String()
	imp.upstreamTarget = cmd.Flag("upstream-target", "set the upstream target for the definition").PlaceHolder("URL").String()
	imp.asMock = cmd.Flag("as-mock", "creates the API as a mock based on example fields").Bool()
	imp.forAPI = cmd.Flag("for-api", "adds blueprint to existing API Definition as version").PlaceHolder("PATH").String()
	imp.asVersion = cmd.Flag("as-version", "the version number to use when inserting").PlaceHolder("VERSION").String()
	// cmd.Action(imp.Import)
}

// Import performs the import process.
// func (i *Importer) Import(ctx *kingpin.ParseContext) (err error) {
// 	if *i.swaggerMode {

// 	}
// }

// func (i *Importer) handleSwaggerMode() error {
// 	if *i.createAPI {
// 		if *i.upstreamTarget != "" && *i.orgID != "" {
// 			// Create the API with the blueprint
// 			//
// 		}
// 	}
// }

// func (i *Importer) swaggerLoadFile(path string) (*importer.swaggerAST, error) {

// }
