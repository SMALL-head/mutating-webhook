package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog"
)

var (
	scheme       = runtime.NewScheme()
	codecs       = serializer.NewCodecFactory(scheme)
	deserializer = codecs.UniversalDeserializer()

	// defaulter = runtime.ObjectDefaulter(scheme)
)

func init() {
	addToScheme(scheme)
}

func addToScheme(scheme *runtime.Scheme) {
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(admissionregistrationv1.AddToScheme(scheme))
	utilruntime.Must(admissionv1.AddToScheme(scheme))
}

// admitv1veta1Func 定义admissionv1beta1.AdmissionReview的处理函数
type admitv1veta1Func func(admissionv1beta1.AdmissionReview) *admissionv1beta1.AdmissionResponse

// admitv1Func 定义admissionv1.AdmissionReview的处理函数
type admitv1Func func(admissionv1.AdmissionReview) *admissionv1.AdmissionResponse

// admitHandler 两种类型的admissionReview的处理函数
type admitHandler struct {
	v1beta1 admitv1veta1Func
	v1      admitv1Func
}

func newAdmitHandlerWithTwoFunc(v1beta1 admitv1veta1Func, v1 admitv1Func) admitHandler {
	return admitHandler{
		v1beta1: v1beta1,
		v1:      v1,
	}
}

func newAdmitHandlerWithV1Func(v1 admitv1Func) admitHandler {
	return newAdmitHandlerWithTwoFunc(convertV1AdmitFuncToV1beta1(v1), v1)
}

// CommandParameter server启动参数
type CommandParameter struct {
	port     int    // server port
	certFile string // path to the x509 certificate for https
	keyFile  string // path to the x509 private key matching `CertFile`
}

type WebhookServer struct {
	server *http.Server
}

func (s *WebhookServer) serve(w http.ResponseWriter, r *http.Request, admit admitHandler) {
	var body []byte
	if r.Body != nil {
		if data, err := io.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	decodeObj, gvk, err := deserializer.Decode(body, nil, nil)
	if err != nil {
		msg := fmt.Sprintf("Request could not be decoded: %v", err)
		klog.Errorf("Can't decode body: %v", err)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	var responseObj runtime.Object
	switch *gvk {
	case admissionv1beta1.SchemeGroupVersion.WithKind("AdmissionReview"):
		requestedAdmissionReview, ok := decodeObj.(*admissionv1beta1.AdmissionReview)
		if !ok {
			msg := fmt.Sprintf("Expected v1beta1.AdmissionReview but got: %T", decodeObj)
			klog.Errorf(msg)
			sendErrorResp(w, msg)
			return
		}
		responseAdmissionReview := &admissionv1beta1.AdmissionReview{}
		responseAdmissionReview.SetGroupVersionKind(*gvk)
		responseAdmissionReview.Response = admit.v1beta1(*requestedAdmissionReview)
		responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
		responseObj = responseAdmissionReview
	case admissionv1.SchemeGroupVersion.WithKind("AdmissionReview"):
		requestedAdmissionReview, ok := decodeObj.(*admissionv1.AdmissionReview)
		if !ok {
			msg := fmt.Sprintf("Expected v1.AdmissionReview but got: %T", decodeObj)
			klog.Errorf(msg)
			sendErrorResp(w, msg)
			return
		}
		responseAdmissionReview := &admissionv1.AdmissionReview{}
		responseAdmissionReview.SetGroupVersionKind(*gvk)
		responseAdmissionReview.Response = admit.v1(*requestedAdmissionReview)
		responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
		responseObj = responseAdmissionReview
	default:
		msg := fmt.Sprintf("Unsupported group version kind: %v", gvk)
		klog.Error(msg)
		sendErrorResp(w, msg)
		return
	}
	klog.Infof("sending response: %v", responseObj)
	responseByte, err := json.Marshal(responseObj)
	if err != nil {
		klog.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(responseByte); err != nil {
		klog.Errorf("Can't write response: %v", err)
	}
}

func (s *WebhookServer) serveMutate(w http.ResponseWriter, r *http.Request) {
	s.serve(w, r, newAdmitHandlerWithV1Func(mutateFunc))
}

func sendErrorResp(w http.ResponseWriter, errMsg string) {
	http.Error(w, errMsg, http.StatusBadRequest)
}
