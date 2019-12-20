/*
Copyright © 2019 Ilya V. Logounov <ilya@logounov.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package clone

import (
	. "github.com/Logunov/heydevops/helpers"
	"github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	"os"
	"os/exec"
	re "regexp"
	"strings"
	"sync"
)

type ConfigStruct struct {
	Logger                    *logrus.Logger
	DryRun                    bool
	ExpandBranches            bool
	GitLabURL                 string
	GitLabAPIURL              string
	Token                     string
	DetectMultiBranchFileName string
	RootRemove                string
	CloneThreadsCount         int
	ListOptionsPerPage        int
	Repos                     SkipCloneStringsStruct
	Branches                  BranchesStruct
}

type SkipCloneStringsStruct struct {
	Skip  []string
	Clone []string
}

type SkipCloneRegexStruct struct {
	Skip  []*re.Regexp
	Clone []*re.Regexp
}

type BranchesStruct struct {
	Prefix string
	Suffix string
	Slash  string
	SkipCloneStringsStruct
}

var (
	config                     *ConfigStruct
	log                        *logrus.Logger
	gitLabClient               *gitlab.Client
	reposSkipCloneRegexList    SkipCloneRegexStruct
	branchesSkipCloneRegexList SkipCloneRegexStruct
	gitMutex                   sync.Mutex
	waitGroup                  sync.WaitGroup
)

func Init(configPtr *ConfigStruct) {
	config = configPtr
	log = config.Logger

	addSlashIfEndWithOutSlash(&config.GitLabURL)
	addSlashIfEndWithOutSlash(&config.GitLabAPIURL)

	log.Trace("Config Dry Run: ", config.DryRun)
	log.Trace("Config GitLabURL: ", config.GitLabURL)
	log.Trace("Config GitLabAPIURL: ", config.GitLabAPIURL)
	log.Trace("Config Token: ", config.Token)
	log.Trace("Config ListOptionsPerPage: ", config.ListOptionsPerPage)
	log.Trace("Config Repos Clone: \n", strings.Join(config.Repos.Clone, "\n"))
	log.Trace("Config Repos Skip: \n", strings.Join(config.Repos.Skip, "\n"))
	log.Trace("Config Branches Prefix: ", config.Branches.Prefix)
	log.Trace("Config Branches Suffix: ", config.Branches.Suffix)
	log.Trace("Config Branches Slash: ", config.Branches.Slash)
	log.Trace("Config Branches Clone: \n", strings.Join(config.Branches.Clone, "\n"))
	log.Trace("Config Branches Skip: \n", strings.Join(config.Branches.Skip, "\n"))

	reposSkipCloneRegexList.Clone = compileSkipCloneRegexps(config.Repos.Clone)
	reposSkipCloneRegexList.Skip = compileSkipCloneRegexps(config.Repos.Skip)
	branchesSkipCloneRegexList.Clone = compileSkipCloneRegexps(config.Branches.Clone)
	branchesSkipCloneRegexList.Skip = compileSkipCloneRegexps(config.Branches.Skip)

	logTraceSkipCloneRegexps("Regexp Repos Cloneinfo", reposSkipCloneRegexList.Clone)
	logTraceSkipCloneRegexps("Regexp Repos Skipinfo", reposSkipCloneRegexList.Skip)
	logTraceSkipCloneRegexps("Regexp Branches Cloneinfo", branchesSkipCloneRegexList.Clone)
	logTraceSkipCloneRegexps("Regexp Branches Skipinfo", branchesSkipCloneRegexList.Skip)

	log.Trace("Core init done")
}

func addSlashIfEndWithOutSlash(strPtr *string) {
	if (*strPtr)[len(*strPtr)-1:] != "/" {
		*strPtr += "/"
	}
}

func compileSkipCloneRegexps(stringsSkipClone []string) []*re.Regexp {
	var regexps []*re.Regexp
	for _, regexpString := range stringsSkipClone {
		regexp, err := re.Compile(regexpString)
		if err != nil {
			log.Error("Can't compile regexp: "+regexpString, err)
			continue
		}
		regexps = append(regexps, regexp)
	}
	return regexps
}

func logTraceSkipCloneRegexps(msg string, regexps []*re.Regexp) {
	for _, regexp := range regexps {
		log.WithFields(logrus.Fields{
			"regexp": regexp.String(),
		}).Trace(msg)
	}
}

func Clone() {
	defer Elapsed("Total")()

	if config.Token == "" {
		log.Fatal("GitLab Token is empty!")
	}

	if config.DryRun {
		log.Info("Running in dry run mode, no really changes will be made")
	}

	gitLabClient = gitlab.NewClient(nil, config.Token)

	if config.GitLabAPIURL == "" {
		config.GitLabAPIURL = config.GitLabURL
	}
	CheckPanic(gitLabClient.SetBaseURL(config.GitLabAPIURL))

	listProjectsOptions := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: config.ListOptionsPerPage,
			Page:    1,
		},
	}

	projectsChan := make(chan *gitlab.Project, config.CloneThreadsCount)
	for i := 0; i < config.CloneThreadsCount; i++ {
		waitGroup.Add(1)
		go addProject(projectsChan, &waitGroup)
	}

	for {
		// Get the first page with projectsChan.
		projects, response, err := gitLabClient.Projects.ListProjects(listProjectsOptions)
		CheckPanic(err)

		// List all the projectsChan we've found so far.
		for _, project := range projects {
			projectsChan <- project
		}

		// Exit the loop when we've seen all pages.
		if response.CurrentPage >= response.TotalPages {
			break
		}

		// Update the page number to get the next page.
		listProjectsOptions.Page = response.NextPage
	}

	close(projectsChan)
	log.Debug("All repos found, now waiting for cloning them ...")
	waitGroup.Wait()
}

func addProject(projectsPtr <-chan *gitlab.Project, waitGroup *sync.WaitGroup) {
	for projectPtr := range projectsPtr {
		repoPath := strings.ReplaceAll(projectPtr.WebURL, config.GitLabURL+config.RootRemove, "")

		if config.RootRemove != "" {
			repoPath = strings.ReplaceAll(repoPath, config.RootRemove, "")
		}

		log.WithFields(logrus.Fields{
			"repo": repoPath,
		}).Debug("project found")

		if !checkSkipCloneRegexps(&reposSkipCloneRegexList, repoPath) {
			log.WithFields(logrus.Fields{
				"repo": repoPath,
			}).Info("repo skipped")
		} else {
			log.WithFields(logrus.Fields{
				"repo": repoPath,
			}).Info("repo clone started")

			// Unused, not needed now
			//if config.DetectMultiBranchFileName != "" {
			//	getFileMetaDataOptions := &gitlab.GetFileMetaDataOptions{
			//		Ref: gitlab.String(projectPtr.DefaultBranch),
			//	}
			//	_, _, err := gitLabClient.RepositoryFiles.GetFileMetaData(projectPtr.ID, config.DetectMultiBranchFileName, getFileMetaDataOptions)
			//	if err == nil {
			//		addMultiBranchRepo(projectPtr.ID, repoPath, projectPtr.SSHURLToRepo)
			//	}
			//}

			if config.ExpandBranches {
				addMultiBranchRepo(repoPath, projectPtr)
			} else {
				addSingleBranchRepo(repoPath, projectPtr.SSHURLToRepo, projectPtr.DefaultBranch, true, "")
			}
		}
	}
	waitGroup.Done()
}

func checkSkipCloneRegexps(regexpsPtr *SkipCloneRegexStruct, str string) bool {
	cloneProject := false

	for _, regexp := range regexpsPtr.Clone {
		if regexp.MatchString(str) {
			cloneProject = true
			log.WithFields(logrus.Fields{
				"regexp": regexp.String(),
				"str":    str,
			}).Trace("matched")
			break
		}
	}

	if !cloneProject {
		log.WithFields(logrus.Fields{
			"str": str,
		}).Trace("didn't match any clone regexp")
		return false
	}

	for _, regexp := range regexpsPtr.Skip {
		if regexp.MatchString(str) {
			log.WithFields(logrus.Fields{
				"regexp": regexp.String(),
				"str":    str,
			}).Trace("skipped due to skip regexp")
			return false
		}
	}

	return true
}

func addMultiBranchRepo(repoPath string, projectPtr *gitlab.Project) {
	listBranchesOptions := &gitlab.ListBranchesOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: config.ListOptionsPerPage,
			Page:    1,
		},
	}

	addSingleBranchRepo(repoPath, projectPtr.SSHURLToRepo, projectPtr.DefaultBranch, true, "")

	for {
		// Get the first page with Branches.
		branches, resp, err := gitLabClient.Branches.ListBranches(projectPtr.ID, listBranchesOptions)
		CheckPanic(err)

		// List all the Branches we've found so far.
		for _, branch := range branches {
			if !branch.Default {
				addSingleBranchRepo(repoPath, projectPtr.SSHURLToRepo, branch.Name, branch.Default, projectPtr.DefaultBranch)
			}
		}

		// Exit the loop when we've seen all pages.
		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		// Update the page number to get the next page.
		listBranchesOptions.Page = resp.NextPage
	}
}

func addSingleBranchRepo(repoPath string, SSHURLToRepo string, branch string, isDefaultBranch bool, defaultBranch string) {
	var branchSlug, branchPath string

	if config.ExpandBranches == true {
		branchSlug = getBranchSlug(branch)
		branchPath = strings.Join([]string{
			repoPath,
			"/",
			config.Branches.Prefix,
			branchSlug,
		}, "")
	} else {
		branchPath = repoPath
	}

	if config.ExpandBranches {
		if !checkSkipCloneRegexps(&branchesSkipCloneRegexList, branch) {
			log.WithFields(logrus.Fields{
				"branch":     branch,
				"branchPath": branchPath,
				"repoPath":   repoPath,
			}).Debug("branch skipped")

			return
		}
	}

	log.WithFields(logrus.Fields{
		"repoPath":      repoPath,
		"branch":        branch,
		"branchPath":    branchPath,
		"defaultBranch": defaultBranch,
	}).Debug("branch clone started")

	_, err := os.Stat(branchPath)
	if os.IsNotExist(err) {
		if isDefaultBranch {
			gitMutex.Lock()
			runCommand("./", "git", "submodule", "add", "--force", "-b", branch, SSHURLToRepo, branchPath)
			gitMutex.Unlock()
		} else {
			defaultBranchPath := repoPath + "/" + config.Branches.Prefix + getBranchSlug(defaultBranch)
			runCommand(defaultBranchPath,
				"git", "worktree", "add", "../"+config.Branches.Prefix+branchSlug, branch)
		}
	}
	runCommand(branchPath, "git", "checkout", branch)
	runCommand(branchPath, "git", "pull")
}

func getBranchSlug(str string) string {
	return strings.ReplaceAll(str, "/", config.Branches.Slash)
}

func runCommand(path string, command string, args ...string) {
	log.WithFields(logrus.Fields{
		"args": args,
		"cmd":  command,
		"path": path,
	}).Trace("runCommand: start")

	if config.DryRun {
		return
	}

	cmd := exec.Command(command, args...)
	cmd.Dir = path

	// And when you need to wait for the command to finish:
	if err := cmd.Run(); err != nil {
		//if err.Error() == "exit status 128" {
		//	log.WithFields(logrus.Fields{
		//		"args": args,
		//		"cmd":  command,
		//		"err":  err,
		//		"path": path,
		//	}).Debug("runCommand: returned error")
		//} else {
		log.WithFields(logrus.Fields{
			"args": args,
			"cmd":  command,
			"err":  err,
			"path": path,
		}).Error("runCommand: returned error")
		//}
	}

	log.WithFields(logrus.Fields{
		"args": args,
		"cmd":  command,
		"path": path,
	}).Trace("runCommand: end")
}
