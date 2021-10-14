package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"gomodules.xyz/jsonpatch/v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	kubermatic "k8c.io/kubermatic/v2/pkg/crd/client/clientset/versioned/typed/kubermatic/v1"

	dmzv1 "github.com/wozniakjan/cluster-wh/api"
	admv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type checkFn func(ar *admv1.AdmissionReview) error
type server struct {
	sv               *http.Server
	kubermaticClient *kubermatic.KubermaticV1Client
	kubeClient       *kubernetes.Clientset
}

var (
	certFile string = "/var/run/secrets/webhook/tls.crt"
	keyFile  string = "/var/run/secrets/webhook/tls.key"
	addr     string = ":9443"
)

func main() {
	flag.StringVar(&certFile, "cert", certFile, "Path to TLS certificate")
	flag.StringVar(&keyFile, "key", keyFile, "Path to TLS certificate key")
	flag.StringVar(&addr, "addr", addr, "Addres to bind the webhook server to")
	flag.Set("logtostderr", "true")
	klog.InitFlags(nil)
	flag.Parse()

	s := newServer()
	if err := s.serve(); err != nil {
		klog.Fatalf("webhook server failed: %v", err)
	}
}

func getPatch(old, updated *admv1.AdmissionReview) ([]byte, *admv1.PatchType, error) {
	klog.V(9).Infof("old raw body: %v", string(old.Request.Object.Raw))
	klog.V(9).Infof("updated raw body: %v", string(updated.Request.Object.Raw))
	patchObj, err := jsonpatch.CreatePatch(old.Request.Object.Raw, updated.Request.Object.Raw)
	if err != nil {
		klog.Errorf("failed to create patch: %v", err)
		return nil, nil, fmt.Errorf("failed to jsonpatch")
	}
	if len(patchObj) == 0 {
		klog.Infof("no patch necessary for request %v/%v", updated.Request.Namespace, updated.Request.Name)
		return nil, nil, nil
	}
	patch, err := json.Marshal(patchObj)
	if err != nil {
		klog.Errorf("failed to marshal JSON patch: %v", err)
		return nil, nil, fmt.Errorf("failed to marshal json")
	}
	pt := admv1.PatchTypeJSONPatch
	klog.Infof("created patch %v, request %v/%v", string(patch), updated.Request.Namespace, updated.Request.Name)
	return patch, &pt, nil
}

func (s *server) mutateCluster(ar *admv1.AdmissionReview) error {
	klog.Infof("defaulting cluster, request %v/%v", ar.Request.Namespace, ar.Request.Name)

	cluster := &dmzv1.DMZCluster{}
	if err := json.Unmarshal(ar.Request.Object.Raw, cluster); err != nil {
		klog.Errorf("failed to unmarshal request %v/%v to cluster: %v", ar.Request.Namespace, ar.Request.Name, err)
		return fmt.Errorf("failed to parse request raw object")
	}

	if err := s.proxyCluster(cluster); err != nil {
		return err
	}

	var err error
	if ar.Request.Object.Raw, err = json.Marshal(cluster); err != nil {
		klog.Errorf("failed to marshal request %v/%v to cluster: %v", ar.Request.Namespace, ar.Request.Name, err)
		return fmt.Errorf("failed to compress request raw object")
	}

	return nil
}

func serve(resp http.ResponseWriter, req *http.Request, check checkFn) {
	klog.Info("serve resp req")
	if req.Body == nil {
		klog.Errorf("req content body is nil")
		return
	}
	contentType := req.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	reqReview := &admv1.AdmissionReview{}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		klog.Errorf("failed reading req body: %v", err)
		return
	}

	if err := json.Unmarshal(body, reqReview); err != nil {
		klog.Infof("failed parsing request body: %v", string(body))
		klog.Errorf("failed parsing request body: %v", err)
		return
	}

	respReview := &admv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: &admv1.AdmissionResponse{
			UID:     reqReview.Request.UID,
			Allowed: true,
			Result: &metav1.Status{
				Status:  "Success",
				Message: "Request permitted",
				Code:    http.StatusOK,
			},
		},
	}
	reqReviewCopy := reqReview.DeepCopy()
	klog.Infof("request raw body: %v", string(reqReview.Request.Object.Raw))
	if err := check(reqReview); err == nil {
		patch, patchType, err := getPatch(reqReviewCopy, reqReview)
		if err != nil {
			klog.Errorf("failed making patch: %v", err)
			return
		}
		r := respReview.Response
		r.Patch, r.PatchType = patch, patchType
	}
	respBytes, err := json.Marshal(respReview)
	if err != nil {
		klog.Errorf("failed to marshal response: %v", err)
		return
	}
	if _, err := resp.Write(respBytes); err != nil {
		klog.Errorf("failed to write response: %v", err)
		return
	} else {
		klog.Infof("response", string(respBytes))
	}
	klog.Infof("finished successfully %v/%v", reqReview.Request.Namespace, reqReview.Request.Name)
}

func (s *server) mutateClusterHandler(resp http.ResponseWriter, req *http.Request) {
	serve(resp, req, s.mutateCluster)
}

func configTLS() *tls.Config {
	sCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		klog.Fatalf("failed to load x509 key pair: %v", err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{sCert},
	}
}

func newServer() *server {
	klog.Info("creating cluster hack webhook")
	router := mux.NewRouter()
	s := &server{
		sv: &http.Server{
			Addr:         addr,
			TLSConfig:    configTLS(),
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		},
	}

	router.HandleFunc("/mutate-cluster", s.mutateClusterHandler)

	cfg, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		panic(err.Error())
	}
	s.kubermaticClient, err = kubermatic.NewForConfig(cfg)
	if err != nil {
		panic(err.Error())
	}
	s.kubeClient, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err.Error())
	}
	s.sv.Handler = router
	return s
}

func (s *server) serve() error {
	return s.sv.ListenAndServeTLS("", "")
}
