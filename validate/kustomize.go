package validate

import (
	"bytes"
	"io"
	"io/fs"
	"os/exec"
	"path/filepath"
)

type WalkFunc func(path string) (stdout io.Writer, stderr io.Writer, err error)

func KustomizeBuild(basePath string) <-chan Carrier {
	return walkPathAndFindKustomizationFileAnRun(basePath)
}

// walkPathAndFindKustomizationFileAnRun walks the given path and finds the kustomization files
// and runs the kustomize build command on the directory containing the kustomization file.
// It returns a channel that will contain the messages from the kustomize build command.
func walkPathAndFindKustomizationFileAnRun(basePath string) <-chan Carrier {
	msgChan := make(chan Carrier)
	filepath.WalkDir(basePath, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			// is a directory so we can skip it
			return nil
		}
		if d.Name() != "kustomization.yaml" && d.Name() != "kustomization.yml" {
			// if the file is not a kustomization file we can skip it
			return nil
		}
		go func(path string) {
			msgChan <- executeKustomize(path)
		}(filepath.Dir(path))
		return nil
	})
	return msgChan
}

// executeKustomize runs the kustomize build command on the given path
// and returns the stdout and stderr writers and an error if any
// occurred during the execution.
func executeKustomize(path string) Carrier {
	stderrWriter := bytes.NewBuffer([]byte{})
	stdoutWriter := bytes.NewBuffer([]byte{})
	cmd := exec.Command("kustomize", "build", "--enable-helm", "--enable-alpha-plugins", path)
	cmd.Stderr = stderrWriter
	cmd.Stdout = stdoutWriter
	err := cmd.Run()
	return Carrier{
		Path:   path,
		Stdout: stdoutWriter.String(),
		Stderr: stderrWriter.String(),
		Err:    err,
	}
}
