package git4go

import (
	"./testutil"
	"testing"
)

func Test_ReadConfig(t *testing.T) {
	testutil.PrepareWorkspace("test_resources/empty_standard_repo/")
	defer testutil.CleanupWorkspace()

	config, _ := NewConfig()
	err := config.AddFile("test_resources/empty_standard_repo/.git/config", ConfigLevelApp, false)

	if err != nil {
		t.Error(err)
	}

	intValue, err := config.LookupInt32("core.repositoryformatversion")
	if err != nil || intValue != 0 {
		t.Error("It should return 0", err, intValue)
	}

	boolValue, err := config.LookupBool("core.bare")
	if err != nil || boolValue != false {
		t.Error("It should return false", err, boolValue)
	}

	boolValue, err = config.LookupBool("core.ignorecase")
	if err != nil || boolValue != true {
		t.Error("It should return true", err, boolValue)
	}

	strValue, err := config.LookupString("core.hideDotFiles")
	if err != nil || strValue != "dotGitOnly" {
		t.Error("It should return dotGitOnly", err, strValue)
	}
}
