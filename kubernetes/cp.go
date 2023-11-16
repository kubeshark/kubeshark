package kubernetes

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/remotecommand"
)

func CopyFromPod(ctx context.Context, provider *Provider, pod v1.Pod, srcPath string, dstPath string) error {
	reader, outStream := io.Pipe()
	cmdArr := []string{"/bin/sh", "-c", "tar", "cf", "-", srcPath}
	req := provider.clientSet.RESTClient().
		Get().
		Prefix([]string{"api", "v1"}...).
		Namespace(pod.Namespace).
		Resource("pods").
		Name(pod.Name).
		SubResource("exec").
		Param("container", "sniffer").
		VersionedParams(&v1.PodExecOptions{
			Command: cmdArr,
			Stdin:   true,
			Stdout:  true,
			Stderr:  true,
			TTY:     false,
		}, metav1.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(&provider.clientConfig, "POST", req.URL())
	if err != nil {
		return err
	}
	go func() {
		defer outStream.Close()
		err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
			Stdin:  os.Stdin,
			Stdout: outStream,
			Stderr: os.Stderr,
			Tty:    false,
		})
		if err != nil {
			log.Error().Err(err).Send()
		}
	}()
	prefix := getPrefix(srcPath)
	prefix = path.Clean(prefix)
	prefix = stripPathShortcuts(prefix)
	dstPath = path.Join(dstPath, path.Base(prefix))
	err = untarAll(reader, dstPath, prefix)
	return err
}

func untarAll(reader io.Reader, destDir, prefix string) error {
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		if !strings.HasPrefix(header.Name, prefix) {
			return fmt.Errorf("tar contents corrupted")
		}

		mode := header.FileInfo().Mode()
		destFileName := filepath.Join(destDir, header.Name[len(prefix):])

		baseName := filepath.Dir(destFileName)
		if err := os.MkdirAll(baseName, 0755); err != nil {
			return err
		}
		if header.FileInfo().IsDir() {
			if err := os.MkdirAll(destFileName, 0755); err != nil {
				return err
			}
			continue
		}

		evaledPath, err := filepath.EvalSymlinks(baseName)
		if err != nil {
			return err
		}

		if mode&os.ModeSymlink != 0 {
			linkname := header.Linkname

			if !filepath.IsAbs(linkname) {
				_ = filepath.Join(evaledPath, linkname)
			}

			if err := os.Symlink(linkname, destFileName); err != nil {
				return err
			}
		} else {
			outFile, err := os.Create(destFileName)
			if err != nil {
				return err
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}
			if err := outFile.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}

func getPrefix(file string) string {
	return strings.TrimLeft(file, "/")
}

func stripPathShortcuts(p string) string {
	newPath := p
	trimmed := strings.TrimPrefix(newPath, "../")

	for trimmed != newPath {
		newPath = trimmed
		trimmed = strings.TrimPrefix(newPath, "../")
	}

	// trim leftover {".", ".."}
	if newPath == "." || newPath == ".." {
		newPath = ""
	}

	if len(newPath) > 0 && string(newPath[0]) == "/" {
		return newPath[1:]
	}

	return newPath
}
