package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cmdrunner"
	"github.com/jenkins-x/jx-helpers/v3/pkg/files"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/cli"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/giturl"
	"github.com/jenkins-x/jx-helpers/v3/pkg/scmhelpers"
	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
)

const (
	headerTemplate = `---
title: %s
linktitle: %s
type: docs
description: "%s"
aliases:
  - %s
---

`
)

var (
	cloneRepositories = os.Getenv("JX_DISABLE_GIT_CLONE") != "true"

	info = termcolor.ColorInfo
)

func main() {
	o := &Options{}
	if len(os.Args) > 1 {
		o.Dir = os.Args[1]
	}
	err := o.Run()
	if err != nil {
		log.Logger().Errorf("failed: %v", err)
		os.Exit(1)
	}
	log.Logger().Infof("completed the plugin generator")
	os.Exit(0)
}

type Options struct {
	scmhelpers.Factory

	Dir           string
	WorkDir       string
	GitClient     gitclient.Interface
	CommandRunner cmdrunner.CommandRunner
}

// Validate validates the setup
func (o *Options) Validate() error {
	if o.CommandRunner == nil {
		o.CommandRunner = cmdrunner.QuietCommandRunner
	}
	if o.GitClient == nil {
		o.GitClient = cli.NewCLIClient("", o.CommandRunner)
	}
	o.GitServerURL = giturl.GitHubURL

	var err error
	if o.ScmClient == nil {
		o.ScmClient, err = o.Factory.Create()
		if err != nil {
			return errors.Wrapf(err, "failed to create Scm client")
		}
	}

	if o.Dir == "" {
		o.Dir = "."
	}
	if o.WorkDir == "" {
		o.WorkDir = filepath.Join(o.Dir, "jx-plugins")
	}
	log.Logger().Infof("using directory %s", info(o.WorkDir))
	err = os.MkdirAll(o.WorkDir, files.DefaultDirWritePermissions)
	if err != nil {
		return errors.Wrapf(err, "failed to create dir %s", o.WorkDir)
	}
	return nil
}

func (o *Options) Run() error {
	err := o.Validate()
	if err != nil {
		return errors.Wrapf(err, "failed to validate options")
	}

	ctx := context.TODO()
	if cloneRepositories {
		err = o.clonePlugins(ctx)
		if err != nil {
			return errors.Wrapf(err, "failed to clone plugins")
		}
	}
	log.Logger().Infof("now generating the plugin CLI docs")

	err = o.generateDocs()
	if err != nil {
		return errors.Wrapf(err, "failed to generate docs")
	}

	log.Logger().Infof("completed")
	return nil
}

func (o *Options) clonePlugins(ctx context.Context) error {
	repos, _, err := o.ScmClient.Repositories.ListOrganisation(ctx, "jenkins-x-plugins", &scm.ListOptions{
		Size: 1000,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to find repositories")
	}

	for _, repo := range repos {
		if repo.Private {
			continue
		}
		if repo.Archived {
			log.Logger().Infof("ignoring archived repository %s", info(repo.Name))
			continue
		}
		if ignoreRepo(repo) {
			log.Logger().Infof("ignoring repository %s", info(repo.Name))
			continue
		}
		err = o.cloneRepository(repo)
		if err != nil {
			return errors.Wrapf(err, "failed to clone repository")
		}
	}
	return nil
}

func ignoreRepo(repo *scm.Repository) bool {
	return !strings.HasPrefix(repo.Name, "jx-")
}

func (o *Options) cloneRepository(repo *scm.Repository) error {
	gitURL := repo.Clone
	if gitURL == "" {
		log.Logger().Warnf("no clone URL for repository %s", repo.Name)
		return nil
	}

	dir, err := filepath.Abs(o.WorkDir)
	if err != nil {
		return errors.Wrapf(err, "failed to get absolute dir of %s", o.WorkDir)
	}

	toDir := filepath.Join(dir, repo.Name)
	err = os.MkdirAll(toDir, files.DefaultDirWritePermissions)
	if err != nil {
		return errors.Wrapf(err, "failed to create dir %s", toDir)
	}

	log.Logger().Infof("cloning plugin %s to %s ", info(repo.Name), info(toDir))
	_, err = gitclient.CloneToDir(o.GitClient, gitURL, toDir)
	if err != nil {
		return errors.Wrapf(err, "failed to clone %s to %s", gitURL, toDir)
	}
	return nil
}

func (o *Options) generateDocs() error {
	log.Logger().Infof("reading %s", info(o.WorkDir))
	fileNames, err := os.ReadDir(o.WorkDir)
	if err != nil {
		return errors.Wrapf(err, "failed to read dir %s", o.WorkDir)
	}

	destRootDir := filepath.Join(o.Dir, "content", "en", "v3", "develop", "reference", "jx")
	for _, f := range fileNames {
		if !f.IsDir() {
			continue
		}
		name := f.Name()

		destPluginDir := filepath.Join(destRootDir, strings.TrimPrefix(name, "jx-"))
		err := os.RemoveAll(destPluginDir)
		if err != nil {
			return errors.Wrapf(err, "failed remove existing plugin documentation %s", destPluginDir)
		}

		pluginDir := filepath.Join(o.WorkDir, name)
		srcDir := filepath.Join(pluginDir, "docs", "cmd")
		nameDotMd := name + ".md"
		path := filepath.Join(srcDir, nameDotMd)
		exists, err := files.FileExists(path)
		if err != nil {
			return errors.Wrapf(err, "failed to check if file exists %s", path)
		}
		if !exists {
			path = filepath.Join(pluginDir, "README.md")
			readmeExist, err := files.FileExists(path)
			if err != nil {
				return errors.Wrapf(err, "failed to check if file exists %s", path)
			}
			path := filepath.Join(pluginDir, "docs", "cmd")
			docsExist, err := files.DirExists(path)
			if err != nil {
				return errors.Wrapf(err, "failed to check if dir exists %s", path)
			}
			log.Logger().Infof("docs %s exists: %v README exists %v", path, docsExist, readmeExist)
			continue
		}

		log.Logger().Infof("found docs %s", info(path))
		mdFiles, err := os.ReadDir(srcDir)
		if err != nil {
			return errors.Wrapf(err, "failed to read %s", srcDir)
		}

		for _, f := range mdFiles {
			mdName := f.Name()
			if f.IsDir() || !strings.HasSuffix(mdName, ".md") {
				continue
			}

			nameWithoutExt := strings.TrimSuffix(mdName, ".md")
			nameWithoutPrefix := "jx " + strings.TrimPrefix(nameWithoutExt, "jx-")
			nameWithoutJX := strings.TrimPrefix(nameWithoutExt, "jx-")
			parts := strings.Split(nameWithoutJX, "_")

			title := strings.ReplaceAll(nameWithoutPrefix, "_", " ")
			linkTitle := parts[len(parts)-1]

			destDir := filepath.Join(destRootDir, filepath.Join(parts...))

			err = os.MkdirAll(destDir, files.DefaultDirWritePermissions)
			if err != nil {
				return errors.Wrapf(err, "failed to create dir %s", destDir)
			}

			destFile := filepath.Join(destDir, "_index.md")
			if mdName != nameDotMd {
				words := strings.Split(linkTitle, " ")
				linkTitle = strings.TrimPrefix(linkTitle, words[0]+" ")
			}

			path = filepath.Join(srcDir, mdName)
			data, err := os.ReadFile(path)
			if err != nil {
				return errors.Wrapf(err, "failed to read file %s", path)
			}
			description := ReadCobraDescription(string(data))
			md := strings.ReplaceAll(string(data), ".md)", ")")

			// lets replace the links to the _index.md page
			indexLink := fmt.Sprintf("[%s](%s)", name, name)
			newIndexLink := fmt.Sprintf("[%s](..)", name)
			md = strings.ReplaceAll(md, indexLink, newIndexLink)

			for _, prefix := range []string{"## ", "[", "```\n"} {
				md = strings.ReplaceAll(md, prefix+name, prefix+strings.ReplaceAll(name, "-", " "))
			}

			if len(parts) > 2 {
				parentName := "jx-" + strings.Join(parts[0:2], "_")
				indexLink = fmt.Sprintf("](%s)", parentName)
				md = strings.ReplaceAll(md, indexLink, "](..)")
			}

			md = WrapExamplesInCodeBlock(md)

			// lets remove the "see also" section
			i := strings.Index(md, "### SEE ALSO")
			if i > 0 {
				md = md[0:i]
			}

			// lets remove the header as its part of the page metadata
			i = strings.Index(md, "### ")
			if i > 0 {
				md = md[i:]
			}

			md += fmt.Sprintf("\n\n### Source\n\n[jenkins-x-plugins/%s](https://github.com/jenkins-x-plugins/%s)\n", name, name)

			alias := nameWithoutExt
			text := fmt.Sprintf(headerTemplate, title, linkTitle, description, alias) + md

			err = os.WriteFile(destFile, []byte(text), files.DefaultFileWritePermissions)
			if err != nil {
				return errors.Wrapf(err, "failed to save file %s", destFile)
			}
		}
	}
	return nil
}

func ReadCobraDescription(md string) string {
	lines := strings.Split(md, "\n")
	startedHeader := false
	buf := &strings.Builder{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "## ") {
			startedHeader = true
			continue
		}
		if strings.HasPrefix(line, "### ") && startedHeader {
			break
		}
		if startedHeader {
			if buf.Len() > 0 {
				buf.WriteString(" ")
			}
			buf.WriteString(line)
		}
	}
	return buf.String()
}

func WrapExamplesInCodeBlock(md string) string {
	lines := strings.Split(md, "\n")
	var answer []string
	examples := false
	addedBlock := false
	ignoreBlock := false
	for _, line := range lines {
		if examples && strings.HasPrefix(line, "### ") {
			examples = false
		}
		if examples {
			if strings.HasPrefix(line, "  ") && !addedBlock {
				if strings.HasPrefix(line, "  `") {
					ignoreBlock = true
				}
				if !ignoreBlock && !strings.HasPrefix(line, "  *") && strings.TrimSpace(line) != "" {
					answer = append(answer, "  ```bash")
					addedBlock = true
				}
			}
		}

		answer = append(answer, line)

		if examples {
			if !strings.HasPrefix(line, "  ") {
				if addedBlock && !ignoreBlock {
					answer = append(answer, "  ```")
				}
				addedBlock = false
				ignoreBlock = false
			}
		}

		if line == "### Examples" || strings.HasPrefix(line, "### Examples") {
			examples = true
		}
	}
	return strings.Join(answer, "\n")
}
