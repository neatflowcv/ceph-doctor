package cephpodman

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/neatflowcv/ceph-doctor/internal/domain"
	"github.com/neatflowcv/porun"
)

const (
	cephImage        = "quay.io/ceph/ceph:v18.2.7"
	cephConfigDir    = "/etc/ceph"
	commandTimeout   = 30 * time.Second
	cleanupTimeout   = 15 * time.Second
	cephConfigFile   = "ceph.conf"
	cephKeyringFile  = "ceph.client.admin.keyring"
	containerCommand = "ceph -s"
	filePerm         = 0o600
)

type CephClient struct{}

var _ domain.CephClient = (*CephClient)(nil)

var errCephStatusExit = errors.New("ceph -s returned non-zero exit status")

func NewCephClient() *CephClient {
	return &CephClient{}
}

func (c *CephClient) Status(ctx context.Context, cluster *domain.Cluster) (*domain.CephStatus, error) {
	host, err := c.resolveHost()
	if err != nil {
		return nil, fmt.Errorf("resolve podman host: %w", err)
	}

	runtimeCtx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()

	runtime, err := c.newRuntime(runtimeCtx, host)
	if err != nil {
		return nil, err
	}

	imageCtx, imageCancel := context.WithTimeout(ctx, commandTimeout)
	defer imageCancel()

	err = runtime.EnsureImageAvailable(imageCtx, cephImage)
	if err != nil {
		return nil, fmt.Errorf("ensure image %s: %w", cephImage, err)
	}

	return c.collectOne(ctx, runtime, cluster)
}

func (c *CephClient) collectOne(
	ctx context.Context,
	runtime porun.Runtime,
	cluster *domain.Cluster,
) (*domain.CephStatus, error) {
	result := newStatusResult()

	configDir, err := prepareConfigDir(cluster)
	if err != nil {
		return nil, err
	}

	defer cleanupTempDir(configDir, &err)

	containerName := fmt.Sprintf(
		"cephdoctor-status-%s-%d",
		sanitizeContainerName(cluster.Name()),
		time.Now().UnixNano(),
	)

	containerID, err := c.createStatusContainer(ctx, runtime, configDir, containerName)
	if err != nil {
		return nil, err
	}

	defer cleanupContainer(ctx, runtime, containerID, &err)

	err = c.startContainer(ctx, runtime, containerID)
	if err != nil {
		return nil, err
	}

	stdout, stderr, exitCode, err := c.execStatus(ctx, runtime, containerID)
	if err != nil {
		return result, err
	}

	result.Stdout = stdout
	result.Stderr = stderr

	if exitCode != 0 {
		return result, fmt.Errorf("%w: %d", errCephStatusExit, exitCode)
	}

	return result, nil
}

func (c *CephClient) createStatusContainer(
	ctx context.Context,
	runtime porun.Runtime,
	configDir, containerName string,
) (string, error) {
	createCtx, createCancel := context.WithTimeout(ctx, commandTimeout)
	defer createCancel()

	containerID, err := runtime.CreateContainer(createCtx, porun.ContainerSpec{
		Name:    containerName,
		Image:   cephImage,
		Command: []string{"sleep", "infinity"},
		Volumes: []string{fmt.Sprintf("%s:%s:ro,Z", configDir, cephConfigDir)},
	})
	if err != nil {
		return "", fmt.Errorf("create container: %w", err)
	}

	return containerID, nil
}

func (c *CephClient) startContainer(ctx context.Context, runtime porun.Runtime, containerID string) error {
	startCtx, startCancel := context.WithTimeout(ctx, commandTimeout)
	defer startCancel()

	err := runtime.StartContainer(startCtx, containerID)
	if err != nil {
		return fmt.Errorf("start container: %w", err)
	}

	return nil
}

func (c *CephClient) execStatus(
	ctx context.Context,
	runtime porun.Runtime,
	containerID string,
) (string, string, int, error) {
	execCtx, execCancel := context.WithTimeout(ctx, commandTimeout)
	defer execCancel()

	stdout, stderr, exitCode, err := runtime.ExecContainer(execCtx, containerID, containerCommand)
	if err != nil {
		return "", "", 0, fmt.Errorf("exec ceph -s: %w", err)
	}

	return stdout, stderr, exitCode, nil
}

func (c *CephClient) resolveHost() (string, error) {
	if host := os.Getenv("CONTAINER_HOST"); host != "" {
		return host, nil
	}

	host, err := porun.DetectPodmanURI()
	if err != nil {
		return "", fmt.Errorf("detect podman URI: %w", err)
	}

	return host, nil
}

func (c *CephClient) newRuntime(ctx context.Context, host string) (*porun.PodmanRuntime, error) {
	runtime, err := porun.NewPodmanRuntime(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("create podman runtime: %w", err)
	}

	return runtime, nil
}

func writeConfigDir(dir string, cluster *domain.Cluster) error {
	err := os.WriteFile(filepath.Join(dir, cephConfigFile), []byte(buildCephConfig(cluster)), filePerm)
	if err != nil {
		return fmt.Errorf("write ceph.conf: %w", err)
	}

	err = os.WriteFile(filepath.Join(dir, cephKeyringFile), []byte(buildKeyring(cluster)), filePerm)
	if err != nil {
		return fmt.Errorf("write keyring: %w", err)
	}

	return nil
}

func prepareConfigDir(cluster *domain.Cluster) (string, error) {
	configDir, err := os.MkdirTemp("", "cephdoctor-status-*")
	if err != nil {
		return "", fmt.Errorf("create temp config dir: %w", err)
	}

	err = writeConfigDir(configDir, cluster)
	if err != nil {
		_ = os.RemoveAll(configDir)

		return "", err
	}

	return configDir, nil
}

func buildCephConfig(cluster *domain.Cluster) string {
	return fmt.Sprintf(
		"[global]\n        mon_host = %s\n",
		strings.Join(cluster.Hosts(), " "),
	)
}

func buildKeyring(cluster *domain.Cluster) string {
	return fmt.Sprintf(
		"[client.admin]\n"+
			"        key = %s\n"+
			"        caps mds = \"allow *\"\n"+
			"        caps mgr = \"allow *\"\n"+
			"        caps mon = \"allow *\"\n"+
			"        caps osd = \"allow *\"\n",
		cluster.Key(),
	)
}

func sanitizeContainerName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return "cluster"
	}

	var builder strings.Builder

	lastDash := false

	for _, r := range name {
		isAlphaNum := r >= 'a' && r <= 'z' || r >= '0' && r <= '9'
		if isAlphaNum {
			builder.WriteRune(r)

			lastDash = false

			continue
		}

		if !lastDash {
			builder.WriteByte('-')

			lastDash = true
		}
	}

	sanitized := strings.Trim(builder.String(), "-")
	if sanitized == "" {
		return "cluster"
	}

	return sanitized
}

func cleanupTempDir(configDir string, resultErr *error) {
	removeErr := os.RemoveAll(configDir)
	if *resultErr == nil && removeErr != nil {
		*resultErr = fmt.Errorf("remove temp config dir: %w", removeErr)
	}
}

func cleanupContainer(
	ctx context.Context,
	runtime porun.Runtime,
	containerID string,
	resultErr *error,
) {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cleanupTimeout)
	defer cancel()

	removeErr := runtime.RemoveContainer(cleanupCtx, containerID)
	if *resultErr == nil && removeErr != nil {
		*resultErr = fmt.Errorf("remove container: %w", removeErr)
	}
}

func newStatusResult() *domain.CephStatus {
	return &domain.CephStatus{
		Stdout: "",
		Stderr: "",
	}
}
