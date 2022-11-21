package carpentry

import (
	"bytes"
	"fmt"
	"go.arcalot.io/imagebuilder/internal/ce_service"
	"go.arcalot.io/imagebuilder/internal/dto"
	"go.arcalot.io/imagebuilder/internal/images"
	"go.arcalot.io/imagebuilder/internal/requirements"
	"go.arcalot.io/log"
	log2 "log"
	"os"
	"path/filepath"
)

func Carpentry(build_img bool, push_img bool, cec ce_service.ContainerEngineService, conf dto.Carpenter, abspath string,
	filenames []string, logger log.Logger,
	pythonCodeStyleChecker func(abspath string, stdout *bytes.Buffer, logger log.Logger) error) (bool, error) {

	meets_reqs := make([]bool, 3)
	basic_reqs, err := requirements.BasicRequirements(filenames, logger)
	if err != nil {
		return false, err
	}
	meets_reqs[0] = basic_reqs
	container_reqs, err := requirements.ContainerfileRequirements(abspath, logger)
	if err != nil {
		return false, err
	}
	meets_reqs[1] = container_reqs
	lang_req, err := requirements.LanguageRequirements(abspath, filenames, conf.Image_Name, conf.Image_Tag, logger,
		pythonCodeStyleChecker)
	if err != nil {
		return false, err
	}
	meets_reqs[2] = lang_req
	all_checks := AllTrue(meets_reqs)
	if !all_checks {
		return false, nil
	}
	if err := images.BuildImage(build_img, all_checks, cec, abspath, conf.Image_Name, conf.Image_Tag, conf.Quay_Img_Exp,
		logger); err != nil {
		return false, err
	}
	for _, registry := range conf.Registries {
		if err := images.PushImage(all_checks, build_img, push_img, cec, conf.Image_Name, conf.Image_Tag,
			registry.Username, registry.Password, registry.Url, registry.Namespace, logger); err != nil {
			logger.Errorf("(%w)", err)
		}
	}
	return true, nil
}

func AllTrue(checks []bool) bool {
	for _, v := range checks {
		if !v {
			return false
		}
	}
	return true
}

func CliCarpentry(build bool, push bool, logger log.Logger, cec_choice string) error {
	conf, err := dto.Unmarshal(logger)
	if err != nil {
		return fmt.Errorf("error in carpentry configuration file (%w)", err)
	}
	cleanpath := filepath.Clean(conf.Project_Filepath)
	abspath, err := filepath.Abs(cleanpath)
	if err != nil {
		return fmt.Errorf("invalid absolute path to project (%w)", err)
	}
	files, err := os.Open(filepath.Clean(abspath))
	if err != nil {
		return fmt.Errorf("error opening project directory (%w)", err)
	}
	filenames, err := files.Readdirnames(0)
	if err != nil {
		return fmt.Errorf("error reading project directory (%w)", err)
	}
	err = files.Close()
	if err != nil {
		return fmt.Errorf("error closing directory at %s (%w)", abspath, err)
	}
	cec, err := ce_service.NewContainerEngineService(cec_choice)
	if err != nil {
		return fmt.Errorf("invalid container engine client %w", err)
	}
	passed_reqs, err := Carpentry(build, push, cec, conf, abspath, filenames,
		logger,
		requirements.Flake8PythonCodeStyle)
	if err != nil {
		return fmt.Errorf("error during carpentry (%w)", err)
	}
	if !passed_reqs {
		log2.Fatalf("failed requirements check, not building: %s %s", conf.Image_Name, conf.Image_Tag)
	}
	return nil
}
