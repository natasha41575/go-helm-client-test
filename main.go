package main

import (
	"bytes"
	"fmt"
	client "github.com/mittwald/go-helm-client"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/repo"
	"os"
)

var settings = cli.New()

// render from a public non-oci helm repo
func templatePublicNonOCIRepo() {
	fmt.Println("templatePublicNonOCIRepo")

	c, _ := client.New(nil)

	chartRepo := repo.Entry{
		Name: "bitnami",
		URL:  "https://charts.bitnami.com/bitnami",
		// Since helm 3.6.1 it is necessary to pass 'PassCredentialsAll = true'.
		PassCredentialsAll: true,
	}

	// Add a chart-repository to the client.
	if err := c.AddOrUpdateChartRepo(chartRepo); err != nil {
		fmt.Println("error adding chart", err.Error())
	}

	chartSpec := client.ChartSpec{
		ChartName:   "bitnami/wordpress",
		Namespace:   "default",
		ReleaseName: "test",
	}

	options := &client.HelmTemplateOptions{
		KubeVersion: &chartutil.KubeVersion{
			Version: "v1.23.10",
			Major:   "1",
			Minor:   "23",
		},
	}

	out, err := c.TemplateChart(&chartSpec, options)
	if err != nil {
		fmt.Println("error rendering chart", err.Error())
	}

	os.WriteFile("wordpress", out, 0666)
}

// render from a private oci registry
func templatePrivateOCIRepo() {
	fmt.Println("templatePrivateOCIRepo")

	c, _ := client.New(nil)
	actionConfig := new(action.Configuration)
	registryClient, err := registry.NewClient()
	if err != nil {
		panic(err)
	}

	actionConfig.RegistryClient = registryClient

	o := &bytes.Buffer{}
	err = action.NewRegistryLogin(actionConfig).Run(o, "https://us-central1-docker.pkg.dev", os.Getenv("USER"), os.Getenv("PASS"))
	if err != nil {
		fmt.Println("error registry login", err.Error())
	}

	chartSpec := client.ChartSpec{
		ChartName:   "oci://us-central1-docker.pkg.dev/disco-haiku-324600/config-sync-test/simple",
		Namespace:   "default",
		ReleaseName: "test",
	}

	options := &client.HelmTemplateOptions{
		KubeVersion: &chartutil.KubeVersion{
			Version: "v1.23.10",
			Major:   "1",
			Minor:   "23",
		},
	}

	out, err := c.TemplateChart(&chartSpec, options)
	if err != nil {
		fmt.Println("error rendering chart", err.Error())
	}

	os.WriteFile("simple", out, 0666)
}

// run `helm show chart` on a private non-oci repo
func showChartPrivateNonOCIRepo() {
	fmt.Println("showChartPrivateNonOCIRepo")

	show := new(action.Show)
	registryClient, err := registry.NewClient()
	if err != nil {
		panic(err)
	}

	show.SetRegistryClient(registryClient)

	show.ChartPathOptions.RepoURL = "https://raw.githubusercontent.com/natasha41575/cs-helm-test/main"
	show.ChartPathOptions.Password = os.Getenv("GH_TOKEN")
	show.ChartPathOptions.Username = "natasha41575"
	show.OutputFormat = action.ShowChart

	cp, err := show.ChartPathOptions.LocateChart("simple", settings)
	if err != nil {
		fmt.Println("couldn't locate chart", err.Error())
	}

	run, err := show.Run(cp)
	if err != nil {
		fmt.Println("couldn't show chart", err.Error())
	}

	os.WriteFile("show", []byte(run), 0666)
}

func main() {
	os.Remove("show")
	os.Remove("simple")
	os.Remove("wordpress")
	showChartPrivateNonOCIRepo()
	templatePublicNonOCIRepo()
	templatePrivateOCIRepo()
}
