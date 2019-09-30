// +build mage

package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/zhiminwen/magetool/shellkit"

	"github.com/magefile/mage/mg"

	"github.com/zhiminwen/magetool/sshkit"
	"github.com/zhiminwen/quote"
)

var master *sshkit.SSHClient
var project, workingDir string

func init() {
	os.Setenv("MAGEFILE_VERBOSE", "true")

	var err error
	master, err = sshkit.NewSSHClient("master.ocp.io.cpak", "22", "root", "", "/Users/wenzm/.ssh/id_rsa")
	if err != nil {
		log.Fatalf("Failed to init master ssh:%v", err)
	}
	project = "knative-demo"
	workingDir = project
}

func T00_init_namespace() {
	cmd := quote.CmdTemplate(`
    oc login -u admin -p {{ .password }}
    oc new-project {{ .project }} || oc project {{ .project }}
    mkdir -p {{ .dir }}
    mkdir -p {{ .dir }}/backend
    mkdir -p {{ .dir }}/frontend
    
  `, map[string]string{
		"password": os.Getenv("OCP_PASSWORD"),
		"project":  project,
		"dir":      workingDir,
	})
	master.Execute(cmd)
}

type Backend mg.Namespace

func (Backend) T01_build_and_push_image() {
	for _, f := range quote.Word(`main.go`) {
		master.Upload("back-end/"+f, workingDir+"/backend/"+f)
	}

	content := quote.Template(quote.HereDoc(`
    FROM golang as builder
    WORKDIR /build
    COPY *go /build

    RUN CGO_ENABLED=0 go build -o backend-server *.go
    
    FROM alpine
    WORKDIR /app
    COPY --from=builder /build/backend-server /app

    CMD ["./backend-server"]
  `), map[string]string{})

	master.Put(content, workingDir+"/backend/Dockerfile")
	cmd := quote.CmdTemplate(`
    cd {{ .dir }}/backend
    buildah build-using-dockerfile -f ./Dockerfile -t {{ .tag }} .

    buildah push --tls-verify=false --creds=anyone:$(oc whoami -t) {{ .tag }} docker://docker-registry-default.apps.ocp.io.cpak/{{ .project }}/{{ .tag }}
 
  `, map[string]string{
		"dir":     workingDir,
		"project": project,
		"tag":     fmt.Sprintf("%s-backend:latest", project),
	})

	master.Execute(cmd)
}

type Front mg.Namespace

func (Front) T01_build_and_push_image() {
	cmd := quote.CmdTemplate(`
    cd front-end
    tar cf - {{ .list }} | ssh root@master.ocp.io.cpak '(cd {{.project}}/frontend; tar xf -)'
  `, map[string]string{
		"list":    "public server src .gitignore babel.config.js package.json vue.config.js yarn.lock",
		"project": project,
	})
	shellkit.ExecuteShell(cmd)

	content := quote.Template(quote.HereDoc(`
    FROM node:lts-alpine
    WORKDIR /app

    COPY . /app
    RUN npm install && npm run build

    CMD ["node", "server/server.js"]

  `), map[string]string{})

	master.Put(content, workingDir+"/frontend/Dockerfile")
	cmd = quote.CmdTemplate(`
    cd {{ .dir }}/frontend
    buildah build-using-dockerfile -f ./Dockerfile -t {{ .tag }} .
    buildah push --tls-verify=false --creds=anyone:$(oc whoami -t) {{ .tag }} docker://docker-registry-default.apps.ocp.io.cpak/{{ .project }}/{{ .tag }}
 
  `, map[string]string{
		"dir":     workingDir,
		"project": project,
		"tag":     fmt.Sprintf("%s-frontend:latest", project),
	})

	master.Execute(cmd)
}

type KService mg.Namespace

func gen_backend_svc(version, color, appVersion string) string {
	content := quote.Template(quote.HereDoc(`
    apiVersion: serving.knative.dev/v1alpha1
    kind: Service
    metadata:
      name: demo-backend
      namespace: {{ .project }}
    spec:
      template:
        metadata:
          name: demo-backend-{{.version}}
          annotations:
            autoscaling.knative.dev/target: "10"
        spec:
          containers:
          - image: docker-registry.default.svc:5000/{{ .project }}/{{ .tag }}@sha256:160f36500e5a38171dd04f78c0eae609f8fb6e4126eaad29a1a39f9bdf79c716
            env:
            - name: APP_PORT
              value: "8080"
            - name: APP_COLOR
              value: {{ .color }}
            - name: APP_VERSION
              value: "{{ .appVersion }}"
  `), map[string]string{
		"tag":        "knative-demo-backend",
		"project":    project,
		"version":    version,
		"color":      color,
		"appVersion": appVersion,
	})

	return content
}

func (KService) T01_deploy_backend_service_blue() {
	content := gen_backend_svc("v1", "blue", "v1.0")
	master.Put(content, workingDir+"/backend.ksvc.yaml")
	cmd := quote.CmdTemplate(`
    cd {{ .dir }}
    kubectl apply -f backend.ksvc.yaml
  `, map[string]string{
		"dir": workingDir,
	})

	master.Execute(cmd)
	//Failed because of the port is in number, not quoted!
}

func (KService) T03_deploy_backend_service_split() {
	content := gen_backend_svc("v2", "green", "v1.1")
	traffic := quote.Template(quote.HereDoc(`
    traffic:
    - tag: current
      revisionName: demo-backend-v1
      percent: 0
    - tag: candidate
      revisionName: demo-backend-v2
      percent: 100
  `), map[string]string{})

	var sb strings.Builder
	sb.WriteString(content)
	for _, line := range strings.Split(traffic, "\n") {
		sb.WriteString(fmt.Sprintf("  %s\n", line))
	}
	master.Put(sb.String(), workingDir+"/backend.ksvc.split.yaml")
	cmd := quote.CmdTemplate(`
	  cd {{ .dir }}
	  kubectl apply -f backend.ksvc.split.yaml
	`, map[string]string{
		"dir": workingDir,
	})

	master.Execute(cmd)
}

func (KService) T02_deploy_frontend_service() {
	content := quote.Template(quote.HereDoc(`
    apiVersion: serving.knative.dev/v1alpha1
    kind: Service
    metadata:
      name: demo-frontend
      namespace: {{ .project }}
    spec:
      template:
        spec:
          containers:
          - image: docker-registry.default.svc:5000/{{ .project }}/{{ .tag }}@sha256:727e6f2bfe50169002b9c845981eccc36e9446a9523766273c7d87a1d36c52ea
            env:
            - name: APP_PORT
              value: "8080"
            - name: APP_REST_SERVER
              value: http://demo-backend.knative-demo.apps.ocp.io.cpak
  `), map[string]string{
		"tag":     "knative-demo-frontend",
		"project": project,
	})
	master.Put(content, workingDir+"/frontend.ksvc.yaml")
	cmd := quote.CmdTemplate(`
    cd {{ .dir }}
    kubectl apply -f frontend.ksvc.yaml
  `, map[string]string{
		"dir": workingDir,
	})

	master.Execute(cmd)
}
