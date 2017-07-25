package main

// To run this package...
// go run gen.go -- --sdk 3.14.16

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	do "gopkg.in/godo.v2"
)

type service struct {
	Name        string
	Fullname    string
	Namespace   string
	Packages    []string
	TaskName    string
	Version     string
	Input       string
	Output      string
	Swagger     string
	SubServices []service
	Modeler     modeler
}

type modeler string

const (
	swagger     modeler = "Swagger"
	compSwagger modeler = "CompositeSwagger"
)

type mapping struct {
	Plane       string
	InputPrefix string
	Services    []service
}

var (
	gopath          = os.Getenv("GOPATH")
	sdkVersion      string
	autorestDir     string
	swaggersDir     string
	deps            do.S
	services        = []*service{}
	servicesMapping = []mapping{
		{
			Plane:       "arm",
			InputPrefix: "arm-",
			Services: []service{
				{
					Name:    "analysisservices",
					Version: "2016-05-16",
				},
				{
					Name:    "apimanagement",
					Version: "2016-07-07",
				},
				{
					Name:    "apimdeployment",
					Version: "2016-07-07",
					Input:   "apimanagement",
				},
				{
					Name:    "authorization",
					Version: "2015-07-01",
				},
				{
					Name:    "batch",
					Version: "2017-01-01",
					Swagger: "BatchManagement",
				},
				{
					Name:    "billing",
					Version: "2017-02-27-preview",
				},
				{
					Name:    "cdn",
					Version: "2016-10-02",
				},
				{
					Name:    "cognitiveservices",
					Version: "2016-02-01-preview",
				},
				{
					Name:    "commerce",
					Version: "2015-06-01-preview",
				},
				{
					Name:    "compute",
					Version: "2016-04-30-preview",
				},
				{
					Name:    "containerservice",
					Version: "2017-01-31",
					Swagger: "containerService",
					Input:   "compute",
				},
				{
					Name:    "containerregistry",
					Version: "2016-06-27-preview",
				},
				{
					Name:    "customer-insights",
					Version: "2017-01-01",
				},
				// {
				// 	Name: "datalake-analytics",
				// 	SubServices: []service{
				// 		{
				// 			Name:    "account",
				// 			Version: "2016-11-01",
				// 		},
				// 	},
				// },
				{
					Name: "datalake-store",
					SubServices: []service{
						{
							Name:    "account",
							Version: "2016-11-01",
						},
					},
				},
				{
					Name:    "devtestlabs",
					Version: "2016-05-15",
					Swagger: "DTL",
				},
				{
					Name:    "disk",
					Version: "2016-04-30-preview",
					Swagger: "disk",
					Input:   "compute",
				},
				{
					Name:    "dns",
					Version: "2016-04-01",
				},
				{
					Name:    "documentdb",
					Version: "2015-04-08",
				},
				{
					Name:    "eventhub",
					Version: "2015-08-01",
					Swagger: "EventHub",
				},
				{
					Name:    "graphrbac",
					Version: "1.6",
					// Composite swagger
				},
				// {
				// 	Name:    "insights",
				// 	// Composite swagger
				// },
				{
					Name:    "intune",
					Version: "2015-01-14-preview",
				},
				{
					Name:    "iothub",
					Version: "2016-02-03",
				},
				{
					Name:    "keyvault",
					Version: "2015-06-01",
				},
				{
					Name:    "logic",
					Version: "2016-06-01",
					// composite swagger
				},
				{
					Name: "machinelearning",
					SubServices: []service{
						{
							Name:    "commitmentplans",
							Version: "2016-05-01-preview",
							Swagger: "commitmentPlans",
							Input:   "machinelearning",
						},
						{
							Name:    "webservices",
							Version: "2016-05-01-preview",
							Input:   "machinelearning",
						},
					},
				},
				{
					Name:    "mediaservices",
					Version: "2015-10-01",
					Swagger: "media",
				},
				{
					Name:    "mobileengagement",
					Version: "2014-12-01",
					Swagger: "mobile-engagement",
				},
				// {
				// 	Name:    "network",
				// 	Version: "2016-09-01",
				// },
				{
					Name:    "networkwatcher",
					Version: "2016-12-01",
					Swagger: "networkWatcher",
					Input:   "network",
				},
				{
					Name:    "notificationhubs",
					Version: "2016-03-01",
				},
				{
					Name:    "operationalinsights",
					Version: "2015-11-01-preview",
					Swagger: "OperationalInsights",
				},
				{
					Name:    "powerbiembedded",
					Version: "2016-01-29",
				},
				{
					Name:    "recoveryservices",
					Version: "2016-06-01",
				},
				// {
				// 	Name:    "recoveryservicesbackup",
				// 	Version: "2016-06-01",
				// 	// composite swagger
				// },
				{
					Name:    "redis",
					Version: "2016-04-01",
				},
				{
					Name: "resources",
					SubServices: []service{
						{
							Name:    "features",
							Version: "2015-12-01",
						},
						{
							Name:    "links",
							Version: "2016-09-01",
						},
						{
							Name:    "locks",
							Version: "2016-09-01",
						},
						{
							Name:    "policy",
							Version: "2016-12-01",
						},
						{
							Name:    "resources",
							Version: "2016-09-01",
						},
						{
							Name:    "subscriptions",
							Version: "2016-06-01",
						},
					},
				},
				{
					Name:    "scheduler",
					Version: "2016-03-01",
				},
				{
					Name:    "search",
					Version: "2015-08-19",
				},
				{
					Name:    "servermanagement",
					Version: "2016-07-01-preview",
				},
				{
					Name:    "servicebus",
					Version: "2015-08-01",
				},
				{
					Name:    "service-map",
					Version: "2015-11-01-preview",
					Swagger: "arm-service-map",
				},
				// {
				// 	Name:    "sql",
				// 	Version: "2014-04-01",
				// 	Swagger: "sql.core",
				// },
				{
					Name:    "storage",
					Version: "2016-12-01",
				},
				{
					Name:    "storageimportexport",
					Version: "2016-11-01",
				},
				{
					Name:    "trafficmanager",
					Version: "2015-11-01",
				},
				{
					Name:    "web",
					Swagger: "compositeWebAppClient",
					Modeler: compSwagger,
				},
			},
		},
		//{
		// Plane:       "dataplane",
		// InputPrefix: "",
		// Services: []service{
		// 	{
		// 		Name:    "batch",
		// 		Version: "2016-07-01.3.1",
		// 		Swagger: "BatchService",
		// 	},
		// 	{
		// 		Name: "insights",
		// 		// composite swagger
		// 	},
		// 	{
		// 		Name:    "keyvault",
		// 		Version: "2015-06-01",
		// 	},
		// 	{
		// 		Name: "search",
		// 		SubServices: []service{
		// 			{
		// 				Name:    "searchindex",
		// 				Version: "2015-02-28",
		// 				Input:   "search",
		// 			},
		// 			{
		// 				Name:    "searchservice",
		// 				Version: "2015-02-28",
		// 				Input:   "search",
		// 			},
		// 		},
		// 	},
		// 	{
		// 		Name:    "servicefabric",
		// 		Version: "2016-01-28",
		// 	},
		// },
		//},
		{
			Plane:       "",
			InputPrefix: "arm-",
			Services: []service{
				// 	{
				// 		Name:    "batch",
				// 		Version: "2016-07-01.3.1",
				// 		Swagger: "BatchService",
				// 	},
				// 	{
				// 		Name: "insights",
				// 		// composite swagger
				// 	},
				{
					Name:    "keyvault",
					Version: "2016-10-01",
				},
				// 	{
				// 		Name: "search",
				// 		SubServices: []service{
				// 			{
				// 				Name:    "searchindex",
				// 				Version: "2015-02-28",
				// 				Input:   "search",
				// 			},
				// 			{
				// 				Name:    "searchservice",
				// 				Version: "2015-02-28",
				// 				Input:   "search",
				// 			},
				// {
				// 	Name: "datalake-analytics",
				// 	SubServices: []service{
				// 		{
				// 			Name:    "catalog",
				// 			Version: "2016-11-01",
				// 		},
				// 		{
				// 			Name:    "job",
				// 			Version: "2016-11-01",
				// 		},
				// 	},
				// 	{
				// 		Name:    "servicefabric",
				// 		Version: "2016-01-28",
				// 	},
			},
		},
		// {
		// 	Plane:       "",
		// 	InputPrefix: "arm-",
		// 	Services: []service{
		// 		{
		// 			Name: "datalake-store",
		// 			SubServices: []service{
		// 				{
		// 					Name:    "filesystem",
		// 					Version: "2016-11-01",
		// 				},
		// 			},
		// 		},
		// {
		// 	Name: "datalake-analytics",
		// 	SubServices: []service{
		// 		{
		// 			Name:    "catalog",
		// 			Version: "2016-11-01",
		// 		},
		// 		{
		// 			Name:    "job",
		// 			Version: "2016-11-01",
		// 		},
		// 	},
		// },
		// 	},
		// },
	}
)

func main() {
	for _, swaggerGroup := range servicesMapping {
		swg := swaggerGroup
		for _, service := range swg.Services {
			s := service
			initAndAddService(&s, swg.InputPrefix, swg.Plane)
		}
	}
	do.Godo(tasks)
}

func initAndAddService(service *service, inputPrefix, plane string) {
	if service.Swagger == "" {
		service.Swagger = service.Name
	}
	packages := append(service.Packages, service.Name)
	service.TaskName = fmt.Sprintf("%s>%s", plane, strings.Join(packages, ">"))
	service.Fullname = filepath.Join(plane, strings.Join(packages, "/"))
	if service.Modeler == compSwagger {
		service.Input = filepath.Join(inputPrefix+strings.Join(packages, "/"), service.Swagger)
	} else {
		if service.Input == "" {
			service.Input = filepath.Join(inputPrefix+strings.Join(packages, "/"), service.Version, "swagger", service.Swagger)
		} else {
			service.Input = filepath.Join(inputPrefix+service.Input, service.Version, "swagger", service.Swagger)
		}
		service.Modeler = swagger
	}
	service.Namespace = filepath.Join("github.com", "Azure", "azure-sdk-for-go", service.Fullname)
	service.Output = filepath.Join(gopath, "src", service.Namespace)

	if service.SubServices != nil {
		for _, subs := range service.SubServices {
			ss := subs
			ss.Packages = append(ss.Packages, service.Name)
			initAndAddService(&ss, inputPrefix, plane)
		}
	} else {
		services = append(services, service)
		deps = append(deps, service.TaskName)
	}
}

func tasks(p *do.Project) {
	p.Task("default", do.S{"setvars", "generate:all", "management"}, nil)
	p.Task("setvars", nil, setVars)
	p.Use("generate", generateTasks)
	p.Use("gofmt", formatTasks)
	p.Use("gobuild", buildTasks)
	p.Use("golint", lintTasks)
	p.Use("govet", vetTasks)
	p.Use("delete", deleteTasks)
	p.Task("management", do.S{"setvars"}, managementVersion)
}

func setVars(c *do.Context) {
	if gopath == "" {
		panic("Gopath not set\n")
	}

	sdkVersion = c.Args.MustString("s", "sdk", "version")
	autorestDir = c.Args.MayString("C:/", "a", "ar", "autorest")
	swaggersDir = c.Args.MayString("C:/", "w", "sw", "swagger")
}

func generateTasks(p *do.Project) {
	addTasks(generate, p)
}

func generate(service *service) {
	fmt.Printf("Generating %s...\n\n", service.Fullname)
	delete(service)

	autorest := exec.Command(
		"autorest",
		"-Input", filepath.Join(swaggersDir, "azure-rest-api-specs", service.Input+".json"),
		"-CodeGenerator", "Go",
		"-Header", "MICROSOFT_APACHE",
		"-Namespace", service.Name,
		"-OutputDirectory", service.Output,
		"-Modeler", string(service.Modeler),
		"-pv", sdkVersion)
	autorest.Dir = filepath.Join(autorestDir, "autorest")
	if err := runner(autorest); err != nil {
		panic(fmt.Errorf("Autorest error: %s", err))
	}

	format(service)
	build(service)
	lint(service)
	vet(service)
}

func deleteTasks(p *do.Project) {
	addTasks(format, p)
}

func delete(service *service) {
	fmt.Printf("Deleting %s...\n\n", service.Fullname)
	err := os.RemoveAll(service.Output)
	if err != nil {
		panic(fmt.Sprintf("Error deleting %s : %s\n", service.Output, err))
	}
}

func formatTasks(p *do.Project) {
	addTasks(format, p)
}

func format(service *service) {
	fmt.Printf("Formatting %s...\n\n", service.Fullname)
	gofmt := exec.Command("gofmt", "-w", service.Output)
	err := runner(gofmt)
	if err != nil {
		panic(fmt.Errorf("gofmt error: %s", err))
	}
}

func buildTasks(p *do.Project) {
	addTasks(build, p)
}

func build(service *service) {
	fmt.Printf("Building %s...\n\n", service.Fullname)
	gobuild := exec.Command("go", "build", service.Namespace)
	err := runner(gobuild)
	if err != nil {
		panic(fmt.Errorf("go build error: %s", err))
	}
}

func lintTasks(p *do.Project) {
	addTasks(lint, p)
}

func lint(service *service) {
	fmt.Printf("Linting %s...\n\n", service.Fullname)
	golint := exec.Command(filepath.Join(gopath, "bin", "golint"), service.Namespace)
	err := runner(golint)
	if err != nil {
		panic(fmt.Errorf("golint error: %s", err))
	}
}

func vetTasks(p *do.Project) {
	addTasks(vet, p)
}

func vet(service *service) {
	fmt.Printf("Vetting %s...\n\n", service.Fullname)
	govet := exec.Command("go", "vet", service.Namespace)
	err := runner(govet)
	if err != nil {
		panic(fmt.Errorf("go vet error: %s", err))
	}
}

func managementVersion(c *do.Context) {
	version("management")
}

func version(packageName string) {
	versionFile := filepath.Join(packageName, "version.go")
	os.Remove(versionFile)
	template := `package %s

var (
	sdkVersion = "%s"
)
`
	data := []byte(fmt.Sprintf(template, packageName, sdkVersion))
	ioutil.WriteFile(versionFile, data, 0644)
}

func addTasks(fn func(*service), p *do.Project) {
	for _, service := range services {
		s := service
		p.Task(s.TaskName, nil, func(c *do.Context) {
			fn(s)
		})
	}
	p.Task("all", deps, nil)
}

func runner(cmd *exec.Cmd) error {
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if stdout.Len() > 0 {
		fmt.Println(stdout.String())
	}
	if stderr.Len() > 0 {
		fmt.Println(stderr.String())
	}
	return err
}
