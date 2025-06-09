package lib

import (
	"encoding/json"
	"fmt"

	embedded "github.com/icipe-official/Data-Abstraction-Platform-Back-End/database"
)

func MetadataModelGet(name string) (map[string]any, error) {
	contentBytes, err := embedded.MetadataModels.ReadFile(fmt.Sprintf("metadata_models/%s.metadata_model.json", name))
	if err != nil {
		return nil, FunctionNameAndError(MetadataModelGet, err)
	}
	var jsonParsed map[string]any
	if err := json.Unmarshal(contentBytes, &jsonParsed); err != nil {
		return nil, FunctionNameAndError(MetadataModelGet, err)
	}
	return jsonParsed, nil
}

func MetadataModelGenJoinKey(prefix string, suffix string) string {
	return fmt.Sprintf("%s_join_%s", prefix, suffix)
}

const (
	METADATA_MODELS_MISC_VERBOSE_RESPONSE string = "verbose_response"
)

func MetadataModelMiscGet(name string) (map[string]any, error) {
	contentBytes, err := embedded.MiscMetadataModels.ReadFile(fmt.Sprintf("metadata_models_misc/%s.metadata_model.json", name))
	if err != nil {
		return nil, FunctionNameAndError(MetadataModelMiscGet, err)
	}
	var jsonParsed map[string]any
	if err := json.Unmarshal(contentBytes, &jsonParsed); err != nil {
		return nil, FunctionNameAndError(MetadataModelMiscGet, err)
	}
	return jsonParsed, nil
}

func MetadataModelMiscVerboseReponseGet(replaceFgPropertes map[string]any, dataMetadataModel map[string]any) (map[string]any, error) {
	verboseResponseMetadataModel, err := MetadataModelGet(METADATA_MODELS_MISC_VERBOSE_RESPONSE)
	if err != nil {
		return nil, err
	}

	return verboseResponseMetadataModel, nil
}
