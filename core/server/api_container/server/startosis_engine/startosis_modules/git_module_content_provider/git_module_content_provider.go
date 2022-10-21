package git_module_content_provider

import (
	"github.com/go-git/go-git/v5"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/mholt/archiver"
	"io"
	"os"
	"path"
)

const (
	moduleDirPermission         = 0755
	temporaryRepoDirPattern     = "tmp-repo-dir-*"
	temporaryArchiveFilePattern = "temp-module-archive-*.tgz"
)

type GitModuleContentProvider struct {
	modulesTmpDir string
	modulesDir    string
}

func NewGitModuleContentProvider(moduleDir string, tmpDir string) *GitModuleContentProvider {
	return &GitModuleContentProvider{
		modulesDir:    moduleDir,
		modulesTmpDir: tmpDir,
	}
}

func (provider *GitModuleContentProvider) GetModuleContents(moduleURL string) (string, error) {
	parsedURL, err := parseGitURL(moduleURL)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while parsing URL '%v'", moduleURL)
	}
	if parsedURL.relativeFilePath == "" {
		return "", stacktrace.NewError("The relative path to file is empty for '%v'", moduleURL)
	}
	pathToFile := path.Join(provider.modulesDir, parsedURL.relativeFilePath)

	// Load the file if it already exists
	contents, err := os.ReadFile(pathToFile)
	if err == nil {
		return string(contents), nil
	}

	// Otherwise Clone It
	err = provider.atomicClone(parsedURL)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while cloning the Git Repo '%v'", parsedURL)
	}

	// Load it after cloning
	contents, err = os.ReadFile(pathToFile)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred in reading contents of the file '%v'", pathToFile)
	}

	return string(contents), nil
}

func (provider *GitModuleContentProvider) StoreModuleContents(moduleId string, moduleTar []byte) (string, error) {
	parsedModuleId, err := parseGitURL(moduleId)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while parsing the moduleId '%v'", moduleId)
	}
	modulePathOnDisk := path.Join(provider.modulesDir, parsedModuleId.relativeRepoPath)
	_, err = os.Stat(modulePathOnDisk)
	if err == nil {
		return "", stacktrace.NewError("Module '%v' already exists on disk, not overwriting", modulePathOnDisk)
	}

	tempFile, err := os.CreateTemp("", temporaryArchiveFilePattern)
	if err != nil {
		return "", stacktrace.NewError("An error occurred while creating temporary file to write compressed '%v' to", moduleId)
	}
	bytesWritten, err := tempFile.Write(moduleTar)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while writing contents of '%v' to '%v'", moduleId, tempFile.Name())
	}
	if bytesWritten != len(moduleTar) {
		return "", stacktrace.NewError("An error occurred while writing contents of '%v' to '%v'", moduleId, tempFile.Name())
	}
	err = archiver.Unarchive(tempFile.Name(), modulePathOnDisk)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while unarchiving '%v' to '%v'", tempFile.Name(), modulePathOnDisk)
	}

	return modulePathOnDisk, nil
}

// atomicClone This first clones to a temporary directory and then moves it
// TODO make this support versioning via tags, commit hashes or branches
func (provider *GitModuleContentProvider) atomicClone(parsedURL *ParsedGitURL) error {
	// First we clone into a temporary directory
	tempRepoDirPath, err := os.MkdirTemp(provider.modulesTmpDir, temporaryRepoDirPattern)
	if err != nil {
		return stacktrace.Propagate(err, "Error creating temporary directory for the repository to be cloned into")
	}
	defer os.RemoveAll(tempRepoDirPath)
	gitClonePath := path.Join(tempRepoDirPath, parsedURL.relativeRepoPath)
	_, err = git.PlainClone(gitClonePath, false, &git.CloneOptions{URL: parsedURL.gitURL, Progress: io.Discard})
	if err != nil {
		return stacktrace.Propagate(err, "Error in cloning git repository '%v' to '%v'", parsedURL.gitURL, gitClonePath)
	}

	// Then we move it into the target directory
	moduleAuthorPath := path.Join(provider.modulesDir, parsedURL.moduleAuthor)
	modulePath := path.Join(provider.modulesDir, parsedURL.relativeRepoPath)
	fileMode, err := os.Stat(moduleAuthorPath)
	if err == nil && !fileMode.IsDir() {
		return stacktrace.Propagate(err, "Expected '%v' to be a directory but it is something else", moduleAuthorPath)
	}
	if err != nil {
		if err = os.Mkdir(moduleAuthorPath, moduleDirPermission); err != nil {
			return stacktrace.Propagate(err, "An error occurred while creating the directory '%v'", moduleAuthorPath)
		}
	}
	if err = os.Rename(gitClonePath, modulePath); err != nil {
		return stacktrace.Propagate(err, "An error occurred while moving module at temporary destination '%v' to final destination '%v'", gitClonePath, modulePath)
	}

	return nil
}
