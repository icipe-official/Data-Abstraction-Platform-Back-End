package initdatabase

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"

	embedded "github.com/icipe-official/Data-Abstraction-Platform-Back-End/database"
	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
)

func (n *CmdInitDatabaseService) ServiceGroupAuthorizationRulesCreate(ctx context.Context) (int, error) {
	entries, err := fs.ReadDir(embedded.GroupAuthorizationRules, "group_authorization_rules")
	if err != nil {
		return 0, fmt.Errorf("read group_authorization_rules directory failed, err: %v", err)
	}

	successfulUpserts := 0
	for _, entry := range entries {
		pathToFile := fmt.Sprintf("group_authorization_rules/%v", entry.Name())
		fileContent, err := embedded.GroupAuthorizationRules.ReadFile(pathToFile)
		if err != nil {
			return successfulUpserts, fmt.Errorf("read file %v failed, err: %v", pathToFile, err)
		}
		newAuthorizationRules := make([]intdoment.GroupAuthorizationRules, 0)
		if err := json.Unmarshal(fileContent, &newAuthorizationRules); err != nil {
			return successfulUpserts, fmt.Errorf("marshal file content %v from json failed, err: %v", entry.Name(), err)
		}

		if noOfRulesInserted, err := n.repo.RepoGroupAuthorizationRulesUpsertMany(ctx, newAuthorizationRules); err != nil {
			return successfulUpserts, fmt.Errorf("insert rule for file content %v failed, err: %v", entry.Name(), err)
		} else {
			successfulUpserts += noOfRulesInserted
		}
	}

	return successfulUpserts, nil
}
