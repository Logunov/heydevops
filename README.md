# Hey, DevOps!

### Synopsis

heydevops clones group from GitLab to local directory

```
heydevops [flags]
```

### Flags

```
      --clone-threads int           Working threads count (default 10)
  -c, --config string               config file (default "./heydevops.yaml")
  -n, --dry-run                     If true, don't do any changes
  -b, --expand-branches             If true, branches will be expanded into git worktrees
  -a, --gitlab-api-url string       GitLab API address if it is located at non-default path
  -u, --gitlab-url string           GitLab address
  -h, --help                        help for heydevops
      --list-options-per-page int   For paginated GitLab API call result sets, the number of results
                                    to include per page (default 10)
  -l, --log-level string            Level of logging:
                                    PANIC, FATAL, ERROR, WARN, INFO, DEBUG, TRACE (default "warn")
  -t, --token string                GitLab token from http://<gitlab>/profile/personal_access_tokens page
```

#### Config file

##### Example configuration file:

```yaml
gitlab-url: https://gitlab.corp/gitlab/
gitlab-api-url: https://gitlab.corp/gitlab-api/
repos:
  clone:
    - ^infrastructure\/
   skip:
    - ^infrastructure\/<SKIP_PATH_REGEXP>\/
expand-branches: true
list-options-per-page: 50
branches:
  prefix: _
  suffix:
  slash: __
  clone:
    - ^develop$
    - ^master$
    - ^release\/.*
    - ^rc\/.*
    - ^v\d*
  skip:
    - <SKIPE_BRANCH_REGEXP>
```

### Environment variables

TODO: Add environment variables description

### Example

#### Preparation

First of all you need an initialized git repository to run this tool. You can use already existing repository or create it:

```shell script
mkdir -p ~/gitlab/group/clone/here && cd $_
git init
# create configuration file
vim heydevops.yaml
```

##### Dry run and pass token via environment variable
```shell script
export HEYDEVOPS_TOKEN="<GITLAB_PERSONAL_ACCESS_TOKEN>"
heydevops -n
```

##### Set log level to info and pass token via flag

```shell script
heydevops -l INFO -t <TOKEN>
```

##### Set log level to debug and save stdout & stderr to log file

```shell script
heydevops -l DEBUG 2>&1 | tee logs/heydevops-$(date "+%Y%m%d%H%M").log
```

#### Completion

To load completion run:

```shell script
source <(heydevops completion $SHELL)
```

```
Usage:
  heydevops completion [command]

Available Commands:
  bash        Generates bash completion scripts
  zsh         Generates zsh completion scripts
```
