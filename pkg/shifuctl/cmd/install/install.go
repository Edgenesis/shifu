package install

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "install shifu and its dependencies",
	Run: func(cmd *cobra.Command, args []string) {
		install()
	},
}

// install installs shifu and its dependencies.
func install() {
	// verify if docker is properly installed.
	if err := verifyDockerInstallation(); err != nil {
		panic(err)
	}

	// verify if kubectl is properly installed.
	if err := verifyKubectlInstallation(); err != nil {
		panic(err)
	}

	// verify if helm is properly installed.
	if err := verifyHelmInstallation(); err != nil {
		panic(err)
	}

	// verify if there's a kubernetes cluster running.
	if err := verifyKubernetesCluster(); err != nil {
		panic(err)
	}
}

// verifyDockerInstallation verifies if docker is properly installed.
func verifyDockerInstallation() error {
	// Run the "docker" command with the "ps" subcommand to get the list of containers
	_, err := exec.Command("docker", "ps").Output()
	if err != nil {
		fmt.Println("\033[1;31mError: Docker not installed properly for the current user. Error running: docker ps\033[0m")
		fmt.Println("\033[1;33mTo install Docker, run: curl -fsSL https://get.docker.com | sh\033[0m")
		return err
	}

	return nil
}

// verifyKubectlInstallation verifies if kubectl is properly installed.
func verifyKubectlInstallation() error {
	// Run the "kubectl" command with the "version" subcommand to get the version of kubectl
	_, err := exec.Command("kubectl", "version").Output()
	if err != nil {
		fmt.Println("\033[1;31mKubectl not installed properly for the current user. Error running: kubectl version\033[0m")
		fmt.Println("\033[1;33mTo install kubectl, follow the instructions at: https://kubernetes.io/docs/tasks/tools/install-kubectl/\033[0m")
		return err
	}

	return nil
}

// verifyHelmInstallation verifies if helm is properly installed.
func verifyHelmInstallation() error {
	// Run the "helm" command with the "version" subcommand to get the version of helm
	_, err := exec.Command("helm", "version").Output()
	if err != nil {
		fmt.Println("\033[1;31mHelm not installed properly for the current user. Error running: helm version\033[0m")
		fmt.Println("\033[1;33mTo install helm, follow the instructions at: https://helm.sh/docs/intro/install/\033[0m")
		return err
	}

	return nil
}

// verifyKubernetesCluster verifies if there's a kubernetes cluster running.
func verifyKubernetesCluster() error {
	// Run the "kubectl" command with the "get" subcommand to get the list of pods
	_, err := exec.Command("kubectl", "get", "pods").Output()
	if err != nil {
		fmt.Println("\033[1;31mError: Kubernetes cluster not running. Error running: kubectl get pods\033[0m")
		fmt.Println("\033[1;33mTo install Kubernetes, follow the instructions at: https://kubernetes.io/docs/setup/\033[0m")
		return err
	}

	return nil
}
