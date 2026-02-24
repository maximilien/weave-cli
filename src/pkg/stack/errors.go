// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"fmt"
	"strings"
)

// EnhanceHelmError adds helpful context to Helm errors
func EnhanceHelmError(err error, chartPath, releaseName string) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()
	helpMsg := ""

	// Common Helm errors and solutions
	if strings.Contains(errMsg, "cannot re-use a name that is still in use") {
		helpMsg = fmt.Sprintf(`
Helm release "%s" already exists.

Solutions:
  1. Uninstall existing release: helm uninstall %s
  2. Use different release name in weave-stack.yaml
  3. Stop stack first: weave stack down`, releaseName, releaseName)
	} else if strings.Contains(errMsg, "has no deployed releases") {
		helpMsg = `
No Helm release found. This is normal for first deployment.`
	} else if strings.Contains(errMsg, "timed out waiting for the condition") {
		helpMsg = `
Helm installation timed out waiting for resources to be ready.

Troubleshooting:
  1. Check pod status: kubectl get pods
  2. Check pod events: kubectl describe pod <pod-name>
  3. Check pod logs: kubectl logs <pod-name>
  4. Increase timeout with longer --timeout value`
	} else if strings.Contains(errMsg, "connection refused") || strings.Contains(errMsg, "unable to connect") {
		helpMsg = `
Cannot connect to Kubernetes cluster.

Troubleshooting:
  1. Check cluster is running: kind get clusters
  2. Verify kubectl context: kubectl config current-context
  3. Restart cluster: weave stack down && weave stack up --runtime kind`
	} else if strings.Contains(errMsg, "not found") && strings.Contains(errMsg, "chart") {
		helpMsg = fmt.Sprintf(`
Helm chart not found at: %s

Solutions:
  1. Ensure templates are copied: ls %s/
  2. Regenerate chart: weave stack up --runtime kind
  3. Check chart structure: helm lint %s`, chartPath, chartPath, chartPath)
	}

	if helpMsg != "" {
		return fmt.Errorf("%w%s", err, helpMsg)
	}

	return err
}

// EnhancePodError adds helpful context to pod errors
func EnhancePodError(err error, podName, context string) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()
	helpMsg := ""

	// Common pod errors and solutions
	if strings.Contains(errMsg, "ImagePullBackOff") || strings.Contains(errMsg, "ErrImagePull") {
		helpMsg = `
Pod cannot pull container image.

Troubleshooting:
  1. Check image name in weave-stack.yaml
  2. Verify image exists: docker pull <image>
  3. Check image pull secrets if using private registry`
	} else if strings.Contains(errMsg, "CrashLoopBackOff") {
		helpMsg = fmt.Sprintf(`
Pod is crashing repeatedly.

Troubleshooting:
  1. Check logs: kubectl --context %s logs %s
  2. Check events: kubectl --context %s describe pod %s
  3. Verify resource limits aren't too low
  4. Check application configuration`, context, podName, context, podName)
	} else if strings.Contains(errMsg, "Pending") {
		helpMsg = `
Pod stuck in Pending state.

Troubleshooting:
  1. Check node resources: kubectl get nodes
  2. Check PVC status: kubectl get pvc
  3. Check events: kubectl describe pod <pod-name>
  4. Ensure storage class exists`
	} else if strings.Contains(errMsg, "not found") {
		helpMsg = fmt.Sprintf(`
Pod not found in cluster.

Troubleshooting:
  1. Check if deployment exists: kubectl --context %s get deployments
  2. Verify stack is deployed: weave stack status
  3. Check namespace: kubectl --context %s get pods --all-namespaces | grep %s`, context, context, podName)
	}

	if helpMsg != "" {
		return fmt.Errorf("%w%s", err, helpMsg)
	}

	return err
}

// EnhanceClusterError adds helpful context to cluster errors
func EnhanceClusterError(err error, provider string) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()
	helpMsg := ""

	// Common cluster errors and solutions
	if strings.Contains(errMsg, "cluster already exists") {
		helpMsg = fmt.Sprintf(`
Cluster already exists.

Solutions:
  1. Use existing cluster: weave stack up --skip-cluster-creation
  2. Delete and recreate: %s delete cluster <name> && weave stack up --runtime %s`, provider, provider)
	} else if strings.Contains(errMsg, "command not found") || strings.Contains(errMsg, "executable file not found") {
		helpMsg = fmt.Sprintf(`
%s not installed.

Installation:
  kind:     brew install kind (macOS) or https://kind.sigs.k8s.io/docs/user/quick-start/
  minikube: brew install minikube (macOS) or https://minikube.sigs.k8s.io/docs/start/`, provider)
	} else if strings.Contains(errMsg, "failed to create cluster") {
		helpMsg = `
Cluster creation failed.

Troubleshooting:
  1. Check container runtime is running: podman machine start
  2. Verify resources available: podman info
  3. Check logs in output above for specific error`
	}

	if helpMsg != "" {
		return fmt.Errorf("%w%s", err, helpMsg)
	}

	return err
}

// CheckDependencies verifies required tools are installed
func CheckDependencies(runtime string) error {
	var missing []string

	// Always required
	if !CommandExists("kubectl") {
		missing = append(missing, "kubectl")
	}
	if !CommandExists("helm") {
		missing = append(missing, "helm")
	}

	// Runtime-specific
	switch runtime {
	case "kind":
		if !CommandExists("kind") {
			missing = append(missing, "kind")
		}
	case "minikube":
		if !CommandExists("minikube") {
			missing = append(missing, "minikube")
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf(`missing required dependencies: %s

Installation:
  kubectl:  brew install kubectl (macOS) or https://kubernetes.io/docs/tasks/tools/
  helm:     brew install helm (macOS) or https://helm.sh/docs/intro/install/
  kind:     brew install kind (macOS) or https://kind.sigs.k8s.io/docs/user/quick-start/
  minikube: brew install minikube (macOS) or https://minikube.sigs.k8s.io/docs/start/

Tip: Run this after installing: weave stack up --runtime %s`, strings.Join(missing, ", "), runtime)
	}

	return nil
}
