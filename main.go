package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const name = "kubectl-wipeout"

const version = "0.0.1"

var revision = "HEAD"

func fatalIf(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func main() {
	var showVersion bool
	flag.BoolVar(&showVersion, "V", false, "Print the version")
	flag.Parse()

	if showVersion {
		fmt.Printf("%s %s (rev: %s/%s)\n", name, version, revision, runtime.Version())
		return
	}

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	fatalIf(err)

	clientset, err := kubernetes.NewForConfig(config)
	fatalIf(err)

	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	fatalIf(err)

	for _, node := range pods.Items {
		if len(node.GetOwnerReferences()) > 0 {
			continue
		}
		ns := node.GetObjectMeta().GetNamespace()
		po := node.GetObjectMeta().GetName()
		if !strings.HasPrefix(po, "kube-proxy") {
			continue
		}
		err = clientset.CoreV1().Pods(ns).Delete(context.TODO(), po, metav1.DeleteOptions{})
		fatalIf(err)
	}
}
