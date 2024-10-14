package main

import (
	"encoding/json"
	"github.com/golang/glog"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	addAllLabelPatch = `
		[ {"op": "add", "path": "/metadata/labels", "value": {"mutated": "yes"} }]
	`
	addAdditionalLabelPatch = `
		[ {"op": "add", "path": "/metadata/labels/mutated", "value": "yes"} ]
	`
	updateLabelPatch = `
		[ {"op": "replace", "path": "/metadata/labels/mutated", "value": "yes"} ]
	`
)

func mutateFunc(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	glog.Info("mutating... 尝试增加一个label")
	obj := struct {
		metav1.ObjectMeta `json:"metadata,omitempty"`
	}{}
	raw := ar.Request.Object.Raw
	err := json.Unmarshal(raw, &obj)
	if err != nil {
		glog.Errorf("json转换失败，err=%s", err)
		return toV1AdmissionErrorResp(err)
	}

	resp := &admissionv1.AdmissionResponse{}
	// mutate操作，所以直接返回true
	resp.Allowed = true

	// 设置resp类型
	// 注意：我们不能直接写成resp.PatchType = &admissionv1.PatchTypeJSONPatch
	// 因为PatchTypeJSONPatch是一个常量，不能取地址
	patchType := admissionv1.PatchTypeJSONPatch
	resp.PatchType = &patchType

	labelValue, hasLabel := obj.ObjectMeta.Labels["mutated"]
	switch {
	case len(obj.ObjectMeta.Labels) == 0: // 缺少标签，因此我们需要把整个标签都加上
		resp.Patch = []byte(addAllLabelPatch)
	case !hasLabel: // 缺少mutated标签，就加上
		resp.Patch = []byte(addAdditionalLabelPatch)
	case labelValue != "yes": // mutated标签已经存在，更新它
		resp.Patch = []byte(updateLabelPatch)
	default:
	}
	
	return resp
}
