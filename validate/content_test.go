package validate

import (
	"testing"

	"github.com/redhat-consulting-services/kustomize-validator/k8s"
)

var (
	stdoutExample = `apiVersion: v1
kind: Pod
metadata:
  name: my-app
  labels:
    app: PATCH_ME
spec:
`
	stdoutExample2 = `apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  selector:
	app: a text with PATCH_ME inside
`
	stdoutExample3 = `apiVersion: v1
kind: ConfigMap
metadata:
  name: config
data:
  image: myapp:latest
  tag: v1.0.0
  debug: true
`
	stdoutExample4 = `apiVersion: v1
kind: Secret
metadata:
  name: credentials
data:
  password: CHANGE_ME_password
  token: CHANGE_ME_token
  apikey: CHANGE_ME_key
`
	stdoutExample5 = `apiVersion: v1
kind: Deployment
metadata:
  name: app
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: app-abc
        image: nginx:latest
        env:
        - name: ENV
          value: dev
`
	stdoutExample6 = `apiVersion: v1
kind: Pod
metadata:
  name: test
  annotations:
    todo: TODO fix this later
    note: FIXME before release
data:
  value: PATCH_ME_NOW
`
	stdoutExample7 = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
spec:
  selector:
    matchLabels:
      app: PATCH_ME_APP
  template:
    metadata:
      labels:
        app: PATCH_ME_APP
`
)

func Test_validateContent(t *testing.T) {
	type args struct {
		resource k8s.Resource
		check    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// Literal matching tests
		{
			name: "simple match",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Pod",
					Name:        "my-app",
					Namespace:   "default",
					SourcePath:  "example/pod.yaml",
					FileContent: stdoutExample,
				},
				check: "PATCH_ME",
			},
			wantErr: true,
		},
		{
			name: "no match",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Pod",
					Name:        "my-app",
					Namespace:   "default",
					SourcePath:  "example/pod.yaml",
					FileContent: stdoutExample,
				},
				check: "abcdef",
			},
			wantErr: false,
		},
		{
			name: "simple match in text",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Service",
					Name:        "my-service",
					Namespace:   "default",
					SourcePath:  "example/service.yaml",
					FileContent: stdoutExample2,
				},
				check: "PATCH_ME",
			},
			wantErr: true,
		},
		{
			name: "no match in text",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Service",
					Name:        "my-service",
					Namespace:   "default",
					SourcePath:  "example/service.yaml",
					FileContent: stdoutExample2,
				},
				check: "abcdef",
			},
			wantErr: false,
		},
		{
			name: "match TODO literal",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Pod",
					Name:        "test",
					Namespace:   "default",
					SourcePath:  "example/pod.yaml",
					FileContent: stdoutExample6,
				},
				check: "TODO",
			},
			wantErr: true,
		},
		{
			name: "match FIXME literal",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Pod",
					Name:        "test",
					Namespace:   "default",
					SourcePath:  "example/pod.yaml",
					FileContent: stdoutExample6,
				},
				check: "FIXME",
			},
			wantErr: true,
		},
		{
			name: "match debug true",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "ConfigMap",
					Name:        "config",
					Namespace:   "default",
					SourcePath:  "example/configmap.yaml",
					FileContent: stdoutExample3,
				},
				check: "debug: true",
			},
			wantErr: true,
		},
		{
			name: "match latest tag",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Deployment",
					Name:        "app",
					Namespace:   "default",
					SourcePath:  "example/deployment.yaml",
					FileContent: stdoutExample5,
				},
				check: ":latest",
			},
			wantErr: true,
		},
		{
			name: "match dev environment",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Deployment",
					Name:        "app",
					Namespace:   "default",
					SourcePath:  "example/deployment.yaml",
					FileContent: stdoutExample5,
				},
				check: "value: dev",
			},
			wantErr: true,
		},
		{
			name: "empty content no match",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Pod",
					Name:        "empty",
					Namespace:   "default",
					SourcePath:  "example/empty.yaml",
					FileContent: "",
				},
				check: "PATCH_ME",
			},
			wantErr: false,
		},
		// Glob pattern tests
		{
			name: "glob match simple",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Pod",
					Name:        "my-app",
					Namespace:   "default",
					SourcePath:  "example/pod.yaml",
					FileContent: stdoutExample,
				},
				check: "glob:PAT*_ME",
			},
			wantErr: true,
		},
		{
			name: "glob match in text",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Service",
					Name:        "my-service",
					Namespace:   "default",
					SourcePath:  "example/service.yaml",
					FileContent: stdoutExample5,
				},
				check: "glob:app-*",
			},
			wantErr: true,
		},
		{
			name: "glob match CHANGE_ME prefix",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Secret",
					Name:        "credentials",
					Namespace:   "default",
					SourcePath:  "example/secret.yaml",
					FileContent: stdoutExample4,
				},
				check: "glob:CHANGE_ME*",
			},
			wantErr: true,
		},
		{
			name: "glob match PATCH_ME with suffix",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Pod",
					Name:        "test",
					Namespace:   "default",
					SourcePath:  "example/pod.yaml",
					FileContent: stdoutExample6,
				},
				check: "glob:PATCH_ME*",
			},
			wantErr: true,
		},
		{
			name: "glob no match",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Pod",
					Name:        "my-app",
					Namespace:   "default",
					SourcePath:  "example/pod.yaml",
					FileContent: stdoutExample,
				},
				check: "glob:NOMATCH*",
			},
			wantErr: false,
		},
		{
			name: "glob match any TODO",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Pod",
					Name:        "test",
					Namespace:   "default",
					SourcePath:  "example/pod.yaml",
					FileContent: stdoutExample6,
				},
				check: "glob:TODO*",
			},
			wantErr: true,
		},
		{
			name: "glob match image tag latest",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Deployment",
					Name:        "app",
					Namespace:   "default",
					SourcePath:  "example/deployment.yaml",
					FileContent: stdoutExample5,
				},
				check: "glob:*:latest",
			},
			wantErr: true,
		},
		// Regex pattern tests
		{
			name: "regex match word boundary PATCH_ME",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Pod",
					Name:        "my-app",
					Namespace:   "default",
					SourcePath:  "example/pod.yaml",
					FileContent: stdoutExample,
				},
				check: `regex:\bPATCH_ME\b`,
			},
			wantErr: true,
		},
		{
			name: "regex match PATCH_ME in text",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Service",
					Name:        "my-service",
					Namespace:   "default",
					SourcePath:  "example/service.yaml",
					FileContent: stdoutExample2,
				},
				check: `regex:\bPATCH_ME\b`,
			},
			wantErr: true,
		},
		{
			name: "regex no match with word boundary on suffix",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Pod",
					Name:        "test",
					Namespace:   "default",
					SourcePath:  "example/pod.yaml",
					FileContent: stdoutExample6,
				},
				check: `regex:\bPATCH_ME\b`,
			},
			wantErr: false, // PATCH_ME_NOW has no word boundary after PATCH_ME
		},
		{
			name: "regex match latest tag with word boundary",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Deployment",
					Name:        "app",
					Namespace:   "default",
					SourcePath:  "example/deployment.yaml",
					FileContent: stdoutExample5,
				},
				check: `regex::latest\b`,
			},
			wantErr: true,
		},
		{
			name: "regex match dev environment",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Deployment",
					Name:        "app",
					Namespace:   "default",
					SourcePath:  "example/deployment.yaml",
					FileContent: stdoutExample5,
				},
				check: `regex:\bdev\b`,
			},
			wantErr: true,
		},
		{
			name: "regex match TODO pattern",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Pod",
					Name:        "test",
					Namespace:   "default",
					SourcePath:  "example/pod.yaml",
					FileContent: stdoutExample6,
				},
				check: `regex:TODO\s+`,
			},
			wantErr: true,
		},
		{
			name: "regex match FIXME pattern",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Pod",
					Name:        "test",
					Namespace:   "default",
					SourcePath:  "example/pod.yaml",
					FileContent: stdoutExample6,
				},
				check: `regex:FIXME\s+`,
			},
			wantErr: true,
		},
		{
			name: "regex match CHANGE_ME with underscore",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Secret",
					Name:        "credentials",
					Namespace:   "default",
					SourcePath:  "example/secret.yaml",
					FileContent: stdoutExample4,
				},
				check: `regex:CHANGE_ME_\w+`,
			},
			wantErr: true,
		},
		{
			name: "regex match any colon latest",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "ConfigMap",
					Name:        "config",
					Namespace:   "default",
					SourcePath:  "example/configmap.yaml",
					FileContent: stdoutExample3,
				},
				check: `regex::latest`,
			},
			wantErr: true,
		},
		{
			name: "regex match debug true",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "ConfigMap",
					Name:        "config",
					Namespace:   "default",
					SourcePath:  "example/configmap.yaml",
					FileContent: stdoutExample3,
				},
				check: `regex:debug:\s*true`,
			},
			wantErr: true,
		},
		{
			name: "regex no match",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Pod",
					Name:        "my-app",
					Namespace:   "default",
					SourcePath:  "example/pod.yaml",
					FileContent: stdoutExample,
				},
				check: `regex:\bNOMATCH\b`,
			},
			wantErr: false,
		},
		{
			name: "regex match PATCH_ME_APP multiple times",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "apps/v1",
					Kind:        "Deployment",
					Name:        "my-deployment",
					Namespace:   "default",
					SourcePath:  "example/deployment.yaml",
					FileContent: stdoutExample7,
				},
				check: `regex:PATCH_ME_APP`,
			},
			wantErr: true,
		},
		{
			name: "regex case sensitive match",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Pod",
					Name:        "my-app",
					Namespace:   "default",
					SourcePath:  "example/pod.yaml",
					FileContent: stdoutExample,
				},
				check: `regex:patch_me`,
			},
			wantErr: false, // Should not match PATCH_ME (case sensitive)
		},
		{
			name: "regex invalid pattern fallback to literal",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Pod",
					Name:        "my-app",
					Namespace:   "default",
					SourcePath:  "example/pod.yaml",
					FileContent: stdoutExample,
				},
				check: `regex:[invalid`,
			},
			wantErr: false, // Invalid regex, falls back to literal which won't match
		},
		// Edge cases
		{
			name: "multiple PATCH_ME on same line",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Pod",
					Name:        "test",
					Namespace:   "default",
					SourcePath:  "example/pod.yaml",
					FileContent: "data: PATCH_ME and PATCH_ME again",
				},
				check: "PATCH_ME",
			},
			wantErr: true,
		},
		{
			name: "pattern at start of line",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Pod",
					Name:        "test",
					Namespace:   "default",
					SourcePath:  "example/pod.yaml",
					FileContent: "PATCH_ME: value",
				},
				check: "PATCH_ME",
			},
			wantErr: true,
		},
		{
			name: "pattern at end of line",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Pod",
					Name:        "test",
					Namespace:   "default",
					SourcePath:  "example/pod.yaml",
					FileContent: "value: PATCH_ME",
				},
				check: "PATCH_ME",
			},
			wantErr: true,
		},
		{
			name: "case sensitive literal",
			args: args{
				resource: k8s.Resource{
					ApiVersion:  "v1",
					Kind:        "Pod",
					Name:        "my-app",
					Namespace:   "default",
					SourcePath:  "example/pod.yaml",
					FileContent: stdoutExample,
				},
				check: "patch_me",
			},
			wantErr: false, // Should not match PATCH_ME
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateContent(tt.args.resource, tt.args.check)
			if (result != nil) != tt.wantErr {
				t.Errorf("expected error: %v, got: %v", tt.wantErr, result)
			}
		})
	}
}
