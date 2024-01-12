package sbom

import (
	"github.com/spdx/spdx-sbom-generator/pkg/runner"
	"github.com/spdx/spdx-sbom-generator/pkg/runner/options"
)

func Create(target string) string {
	opts := options.Options{
		SchemaVersion: "2.3",
		Indent:        4,
		Version:       "2.3",
		License:       false,
		Depth:         "",
		Slug:          "",
		OutputDir:     "/tmp",
		Format:        options.OutputFormatSpdx,
		// GlobalSettingFile: globalSettingFile,
		Path:    ".",
		Plugins: options.DefaultPlugins,
	}

	err := runner.NewWithOptions(opts).CreateSBOM()
	if err != nil {
		panic(err)
	}
	return ""
	// return out.String()
}
