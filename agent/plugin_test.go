package agent

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/buildkite/agent/env"
	"github.com/stretchr/testify/assert"
)

func TestCreatePluginsFromJSON(t *testing.T) {
	var plugins []*Plugin
	var err error

	plugins, err = CreatePluginsFromJSON(`[{"http://github.com/buildkite/plugins/docker-compose#a34fa34":{"container":"app"}}, "github.com/buildkite/plugins/ping#master"]`)
	assert.Equal(t, len(plugins), 2)
	assert.Nil(t, err)

	assert.Equal(t, plugins[0].Location, "github.com/buildkite/plugins/docker-compose")
	assert.Equal(t, plugins[0].Version, "a34fa34")
	assert.Equal(t, plugins[0].Scheme, "http")
	assert.Equal(t, plugins[0].Configuration, map[string]interface{}{"container": "app"})

	assert.Equal(t, plugins[1].Location, "github.com/buildkite/plugins/ping")
	assert.Equal(t, plugins[1].Version, "master")
	assert.Equal(t, plugins[1].Scheme, "")
	assert.Equal(t, plugins[1].Configuration, map[string]interface{}{})

	plugins, err = CreatePluginsFromJSON(`["ssh://git:foo@github.com/buildkite/plugins/docker-compose#a34fa34"]`)
	assert.Equal(t, len(plugins), 1)
	assert.Nil(t, err)

	assert.Equal(t, plugins[0].Location, "github.com/buildkite/plugins/docker-compose")
	assert.Equal(t, plugins[0].Version, "a34fa34")
	assert.Equal(t, plugins[0].Scheme, "ssh")
	assert.Equal(t, plugins[0].Authentication, "git:foo")

	plugins, err = CreatePluginsFromJSON(`blah`)
	assert.Equal(t, len(plugins), 0)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "invalid character 'b' looking for beginning of value")

	plugins, err = CreatePluginsFromJSON(`{"foo": "bar"}`)
	assert.Equal(t, len(plugins), 0)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "JSON structure was not an array")

	plugins, err = CreatePluginsFromJSON(`["github.com/buildkite/plugins/ping#master#lololo"]`)
	assert.Equal(t, len(plugins), 0)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Too many #'s in \"github.com/buildkite/plugins/ping#master#lololo\"")
}

func TestPluginName(t *testing.T) {
	var plugin *Plugin

	plugin = &Plugin{Location: "github.com/buildkite-plugins/docker-compose-buildkite-plugin.git"}
	assert.Equal(t, "docker-compose", plugin.Name())

	plugin = &Plugin{Location: "github.com/buildkite-plugins/docker-compose-buildkite-plugin"}
	assert.Equal(t, "docker-compose", plugin.Name())

	plugin = &Plugin{Location: "github.com/my-org/docker-compose-buildkite-plugin"}
	assert.Equal(t, "docker-compose", plugin.Name())

	plugin = &Plugin{Location: "github.com/buildkite/plugins/docker-compose"}
	assert.Equal(t, "docker-compose", plugin.Name())

	plugin = &Plugin{Location: "github.com/buildkite/my-plugin"}
	assert.Equal(t, "my-plugin", plugin.Name())

	plugin = &Plugin{Location: "~/Development/plugins/test"}
	assert.Equal(t, "test", plugin.Name())

	plugin = &Plugin{Location: "~/Development/plugins/UPPER     CASE_party"}
	assert.Equal(t, "upper-case-party", plugin.Name())

	plugin = &Plugin{Location: "vendor/src/vendored with a space"}
	assert.Equal(t, "vendored-with-a-space", plugin.Name())

	plugin = &Plugin{Location: ""}
	assert.Equal(t, "", plugin.Name())
}

func TestIdentifier(t *testing.T) {
	var plugin *Plugin
	var id string
	var err error

	plugin = &Plugin{Location: "github.com/buildkite/plugins/docker-compose/beta#master"}
	id, err = plugin.Identifier()
	assert.Equal(t, id, "github-com-buildkite-plugins-docker-compose-beta-master")
	assert.Nil(t, err)

	plugin = &Plugin{Location: "github.com/buildkite/plugins/docker-compose/beta"}
	id, err = plugin.Identifier()
	assert.Equal(t, id, "github-com-buildkite-plugins-docker-compose-beta")
	assert.Nil(t, err)

	plugin = &Plugin{Location: "192.168.0.1/foo.git#12341234"}
	id, err = plugin.Identifier()
	assert.Equal(t, id, "192-168-0-1-foo-git-12341234")
	assert.Nil(t, err)

	plugin = &Plugin{Location: "/foo/bar/"}
	id, err = plugin.Identifier()
	assert.Equal(t, id, "foo-bar")
	assert.Nil(t, err)
}

func TestRepositoryAndSubdirectory(t *testing.T) {
	var plugin *Plugin
	var repo string
	var sub string
	var err error

	plugin = &Plugin{Location: "github.com/buildkite/plugins/docker-compose/beta"}
	repo, err = plugin.Repository()
	assert.Equal(t, repo, "https://github.com/buildkite/plugins")
	assert.Nil(t, err)
	sub, err = plugin.RepositorySubdirectory()
	assert.Equal(t, sub, "docker-compose/beta")
	assert.Nil(t, err)

	plugin = &Plugin{Location: "github.com/buildkite/test-plugin"}
	repo, err = plugin.Repository()
	assert.Equal(t, repo, "https://github.com/buildkite/test-plugin")
	assert.Nil(t, err)
	sub, err = plugin.RepositorySubdirectory()
	assert.Equal(t, sub, "")
	assert.Nil(t, err)

	plugin = &Plugin{Location: "github.com/buildkite"}
	repo, err = plugin.Repository()
	assert.Equal(t, repo, "")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), `Incomplete github.com path "github.com/buildkite"`)
	sub, err = plugin.RepositorySubdirectory()
	assert.Equal(t, sub, "")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), `Incomplete github.com path "github.com/buildkite"`)

	plugin = &Plugin{Location: "bitbucket.org/buildkite"}
	repo, err = plugin.Repository()
	assert.Equal(t, repo, "")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), `Incomplete bitbucket.org path "bitbucket.org/buildkite"`)
	sub, err = plugin.RepositorySubdirectory()
	assert.Equal(t, sub, "")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), `Incomplete bitbucket.org path "bitbucket.org/buildkite"`)

	plugin = &Plugin{Location: "bitbucket.org/user/project/sub/directory"}
	repo, err = plugin.Repository()
	assert.Equal(t, repo, "https://bitbucket.org/user/project")
	assert.Nil(t, err)
	sub, err = plugin.RepositorySubdirectory()
	assert.Equal(t, sub, "sub/directory")
	assert.Nil(t, err)

	plugin = &Plugin{Location: "bitbucket.org/user/project/sub/directory", Scheme: "http", Authentication: "foo:bar"}
	repo, err = plugin.Repository()
	assert.Equal(t, repo, "http://foo:bar@bitbucket.org/user/project")
	assert.Nil(t, err)
	sub, err = plugin.RepositorySubdirectory()
	assert.Equal(t, sub, "sub/directory")
	assert.Nil(t, err)

	plugin = &Plugin{Location: "114.135.234.212/foo.git"}
	repo, err = plugin.Repository()
	assert.Equal(t, repo, "https://114.135.234.212/foo.git")
	assert.Nil(t, err)
	sub, err = plugin.RepositorySubdirectory()
	assert.Equal(t, sub, "")
	assert.Nil(t, err)

	plugin = &Plugin{Location: "github.com/buildkite/plugins/docker-compose/beta"}
	repo, err = plugin.Repository()
	assert.Equal(t, repo, "https://github.com/buildkite/plugins")
	assert.Nil(t, err)
	sub, err = plugin.RepositorySubdirectory()
	assert.Equal(t, sub, "docker-compose/beta")
	assert.Nil(t, err)

	plugin = &Plugin{Location: "/Users/keithpitt/Development/plugins.git/test-plugin"}
	repo, err = plugin.Repository()
	assert.Equal(t, repo, "/Users/keithpitt/Development/plugins.git")
	assert.Nil(t, err)
	sub, err = plugin.RepositorySubdirectory()
	assert.Equal(t, sub, "test-plugin")
	assert.Nil(t, err)

	plugin = &Plugin{Location: ""}
	repo, err = plugin.Repository()
	assert.Equal(t, repo, "")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Missing plugin location")
}

func TestConfigurationToEnvironment(t *testing.T) {
	var envMap *env.Environment
	var err error

	envMap, err = pluginEnvFromConfig(t, `{ "config-key": 42 }`)
	assert.Nil(t, err)
	assert.Equal(t, []string{"BUILDKITE_PLUGIN_DOCKER_COMPOSE_CONFIG_KEY=42"}, envMap.ToSlice())

	envMap, err = pluginEnvFromConfig(t, `{ "container": "app", "some-other-setting": "else right here" }`)
	assert.Nil(t, err)
	assert.Equal(t, []string{
		"BUILDKITE_PLUGIN_DOCKER_COMPOSE_CONTAINER=app",
		"BUILDKITE_PLUGIN_DOCKER_COMPOSE_SOME_OTHER_SETTING=else right here"},
		envMap.ToSlice())

	envMap, err = pluginEnvFromConfig(t, `{ "and _ with a    - number": 12 }`)
	assert.Nil(t, err)
	assert.Equal(t, []string{"BUILDKITE_PLUGIN_DOCKER_COMPOSE_AND_WITH_A_NUMBER=12"}, envMap.ToSlice())

	envMap, err = pluginEnvFromConfig(t, `{ "bool-true-key": true, "bool-false-key": false }`)
	assert.Nil(t, err)
	assert.Equal(t, []string{
		"BUILDKITE_PLUGIN_DOCKER_COMPOSE_BOOL_FALSE_KEY=false",
		"BUILDKITE_PLUGIN_DOCKER_COMPOSE_BOOL_TRUE_KEY=true"},
		envMap.ToSlice())

	envMap, err = pluginEnvFromConfig(t, `{ "array-key": [ "array-val-1", "array-val-2" ] }`)
	assert.Nil(t, err)
	assert.Equal(t, []string{
		"BUILDKITE_PLUGIN_DOCKER_COMPOSE_ARRAY_KEY_0=array-val-1",
		"BUILDKITE_PLUGIN_DOCKER_COMPOSE_ARRAY_KEY_1=array-val-2"},
		envMap.ToSlice())

	envMap, err = pluginEnvFromConfig(t, `{ "array-key": [ 42, 43, 44 ] }`)
	assert.Nil(t, err)
	assert.Equal(t, []string{
		"BUILDKITE_PLUGIN_DOCKER_COMPOSE_ARRAY_KEY_0=42",
		"BUILDKITE_PLUGIN_DOCKER_COMPOSE_ARRAY_KEY_1=43",
		"BUILDKITE_PLUGIN_DOCKER_COMPOSE_ARRAY_KEY_2=44"},
		envMap.ToSlice())

	envMap, err = pluginEnvFromConfig(t, `{ "array-key": [ 42, 43, "foo" ] }`)
	assert.Nil(t, err)
	assert.Equal(t, []string{
		"BUILDKITE_PLUGIN_DOCKER_COMPOSE_ARRAY_KEY_0=42",
		"BUILDKITE_PLUGIN_DOCKER_COMPOSE_ARRAY_KEY_1=43",
		"BUILDKITE_PLUGIN_DOCKER_COMPOSE_ARRAY_KEY_2=foo"},
		envMap.ToSlice())
}

func pluginEnvFromConfig(t *testing.T, configJson string) (*env.Environment, error) {
	var config map[string]interface{}

	json.Unmarshal([]byte(configJson), &config)

	jsonString := fmt.Sprintf(`[ { "%s": %s } ]`, "github.com/buildkite-plugins/docker-compose-buildkite-plugin", configJson)

	plugins, err := CreatePluginsFromJSON(jsonString)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(plugins))

	return plugins[0].ConfigurationToEnvironment()
}
