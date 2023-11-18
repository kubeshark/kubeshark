package kubernetes

import (
	"archive/tar"
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

func CopyFromPod(ctx context.Context, provider *Provider, pod v1.Pod, srcPath string, dstPath string) error {
	const containerName = "sniffer"
	cmdArr := []string{"tar", "cf", "-", srcPath}
	req := provider.clientSet.CoreV1().RESTClient().
		Post().
		Namespace(pod.Namespace).
		Resource("pods").
		Name(pod.Name).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: containerName,
			Command:   cmdArr,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(&provider.clientConfig, "POST", req.URL())
	if err != nil {
		return err
	}

	reader, outStream := io.Pipe()
	errReader, errStream := io.Pipe()
	go logErrors(errReader, pod)
	go func() {
		defer outStream.Close()
		err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
			Stdin:  os.Stdin,
			Stdout: outStream,
			Stderr: errStream,
			Tty:    false,
		})
		if err != nil {
			log.Error().Err(err).Str("pod", pod.Name).Msg("SPDYExecutor:")
		}
	}()

	prefix := getPrefix(srcPath)
	prefix = path.Clean(prefix)
	prefix = stripPathShortcuts(prefix)
	dstPath = path.Join(dstPath, path.Base(prefix))
	err = untarAll(reader, dstPath, prefix)
	// fo(reader)
	return err
}

// func fo(fi io.Reader) {
// 	fo, err := os.Create("output.tar")
// 	if err != nil {
// 		panic(err)
// 	}

// 	// make a buffer to keep chunks that are read
// 	buf := make([]byte, 1024)
// 	for {
// 		// read a chunk
// 		n, err := fi.Read(buf)
// 		if err != nil && err != io.EOF {
// 			panic(err)
// 		}
// 		if n == 0 {
// 			break
// 		}

// 		// write a chunk
// 		if _, err := fo.Write(buf[:n]); err != nil {
// 			panic(err)
// 		}
// 	}
// }

func logErrors(reader io.Reader, pod v1.Pod) {
	r := bufio.NewReader(reader)
	for {
		msg, _, err := r.ReadLine()
		log.Warn().Str("pod", pod.Name).Str("msg", string(msg)).Msg("SPDYExecutor:")
		if err != nil {
			if err != io.EOF {
				log.Error().Err(err).Send()
			}
			return
		}
	}
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
