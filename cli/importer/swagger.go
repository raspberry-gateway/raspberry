package importer

import (
	"encoding/json"
	"io"
)

const SwaggerSource APIImporterSource = "swagger"

type DefinitionObjectFormatAST struct {
	Format string `json:"format"`
	Type   string `json:"type"`
}

type DefinitionObjectAST struct {
	Type       string                               `json:"type"`
	Required   []string                             `json:"required"`
	Properties map[string]DefinitionObjectFormatAST `json:"properties"`
}

type ResponseCodeObjectAST struct {
	Description string `json:"description"`
	Schema      struct {
		Items map[string]interface{} `json:"items"`
		Type  string                 `json:"type"`
	}
}

type PathMethodObject struct {
	Description string                           `json:"description"`
	OperationID string                           `json:"operationId"`
	Responses   map[string]ResponseCodeObjectAST `json:"responses"`
}

type PathItemObject struct {
	Get     PathMethodObject `json:"get"`
	Put     PathMethodObject `json:"put"`
	Post    PathMethodObject `json:"post"`
	Patch   PathMethodObject `json:"patch"`
	Options PathMethodObject `json:"options"`
	Delete  PathMethodObject `json:"delete"`
	Head    PathMethodObject `json:"head"`
}

type SwaggerAST struct {
	BasePath    string                         `json:"base_path"`
	Consumes    []string                       `json:"consumes"`
	Definitions map[string]DefinitionObjectAST `json:"definitions"`
	Host        string                         `json:"host"`
	Info        struct {
		Contact struct {
			Email string `json:"email"`
			Name  string `json:"name"`
			URL   string `json:"url"`
		} `json:"contact"`
		Description string `json:"description"`
		License     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"license"`
		TermsOfService string `json:"termsOfService"`
		Title          string `json:"title"`
		Version        string `json:"version"`
	} `json:"info"`
	Paths    map[string]PathItemObject `json:"paths"`
	Produces []string                  `json:"produces"`
	Schemes  []string                  `json:"schemes"`
	Swagger  string                    `json:"swagger"`
}

func (s *SwaggerAST) LoadFrom(r io.Reader) error {
	return json.NewDecoder(r).Decode(&s)
}

func (s *SwaggerAST) ConvertIntoApiVersion(asMock bool) {

}
