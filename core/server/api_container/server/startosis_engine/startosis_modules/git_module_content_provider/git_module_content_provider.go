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
	defaultTmpDir               = ""
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

func (provider *GitModuleContentProvider) GetOnDiskAbsoluteFilePath(fileInsideModuleUrl string) (string, error) {
	parsedURL, err := parseGitURL(fileInsideModuleUrl)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while parsing URL '%v'", fileInsideModuleUrl)
	}
	if parsedURL.relativeFilePath == "" {
		return "", stacktrace.NewError("The relative path to file is empty for '%v'", fileInsideModuleUrl)
	}
	pathToFile := path.Join(provider.modulesDir, parsedURL.relativeFilePath)

	// Return the file path straight if it exists
	if _, err := os.Stat(pathToFile); err == nil {
		return pathToFile, nil
	}

	// Otherwise clone the repo and return the absolute path of the requested file
	err = provider.atomicClone(parsedURL)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while cloning the Git Repo '%v'", parsedURL)
	}
	return pathToFile, nil
}

func (provider *GitModuleContentProvider) GetModuleContents(fileInsideModuleUrl string) (string, error) {
	pathToFile, err := provider.GetOnDiskAbsoluteFilePath(fileInsideModuleUrl)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred loading the module file '%v'", fileInsideModuleUrl)
	}

	// Load the file content from its absolute path
	contents, err := os.ReadFile(pathToFile)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred in reading contents of the file '%v'", pathToFile)
	}

	return string(contents), nil
}

func (provider *GitModuleContentProvider) StoreModuleContents(moduleId string, moduleTar []byte, overwriteExisting bool) (string, error) {
	parsedModuleId, err := parseGitURL(moduleId)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while parsing the module ID '%v'", moduleId)
	}
	modulePathOnDisk := path.Join(provider.modulesDir, parsedModuleId.relativeRepoPath)

	if overwriteExisting {
		err = os.RemoveAll(modulePathOnDisk)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred while removing the existing module '%v' from disk at '%v'", moduleId, modulePathOnDisk)
		}
	}

	_, err = os.Stat(modulePathOnDisk)
	if err == nil {
		return "", stacktrace.NewError("Module '%v' already exists on disk, not overwriting", modulePathOnDisk)
	}

	tempFile, err := os.CreateTemp(defaultTmpDir, temporaryArchiveFilePattern)
	if err != nil {
		return "", stacktrace.NewError("An error occurred while creating temporary file to write compressed '%v' to", moduleId)
	}
	defer os.Remove(tempFile.Name())

	bytesWritten, err := tempFile.Write(moduleTar)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while writing contents of '%v' to '%v'", moduleId, tempFile.Name())
	}
	if bytesWritten != len(moduleTar) {
		return "", stacktrace.NewError("Expected to write '%v' bytes but wrote '%v'", len(moduleTar), bytesWritten)
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
